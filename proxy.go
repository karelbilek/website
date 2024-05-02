package main

// (C) 2022 Drew Devault, see https://git.sr.ht/~sircmpwn/kineto
import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	_ "embed"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/karelbilek/website/logs"
)

var gemtextPage = template.Must(template.
	New("gemtext").
	Funcs(template.FuncMap{
		"heading": func(line gemini.Line) *GemtextHeading {
			switch l := line.(type) {
			case gemini.LineHeading1:
				return &GemtextHeading{1, string(l), createAnchor(string(l))}
			case gemini.LineHeading2:
				return &GemtextHeading{2, string(l), createAnchor(string(l))}
			case gemini.LineHeading3:
				return &GemtextHeading{3, string(l), createAnchor(string(l))}
			default:
				return nil
			}
		},
		"link": func(line gemini.Line) *gemini.LineLink {
			switch l := line.(type) {
			case gemini.LineLink:
				return &l
			default:
				return nil
			}
		},
		"li": func(line gemini.Line) *gemini.LineListItem {
			switch l := line.(type) {
			case gemini.LineListItem:
				return &l
			default:
				return nil
			}
		},
		"pre_toggle_on": func(ctx *GemtextContext, line gemini.Line) *gemini.LinePreformattingToggle {
			switch l := line.(type) {
			case gemini.LinePreformattingToggle:
				if ctx.Pre%4 == 0 {
					ctx.Pre += 1
					return &l
				}
				ctx.Pre += 1
				return nil
			default:
				return nil
			}
		},
		"pre_toggle_off": func(ctx *GemtextContext, line gemini.Line) *gemini.LinePreformattingToggle {
			switch l := line.(type) {
			case gemini.LinePreformattingToggle:
				if ctx.Pre%4 == 3 {
					ctx.Pre += 1
					return &l
				}
				ctx.Pre += 1
				return nil
			default:
				return nil
			}
		},
		"pre": func(line gemini.Line) *gemini.LinePreformattedText {
			switch l := line.(type) {
			case gemini.LinePreformattedText:
				return &l
			default:
				return nil
			}
		},
		"quote": func(line gemini.Line) *gemini.LineQuote {
			switch l := line.(type) {
			case gemini.LineQuote:
				return &l
			default:
				return nil
			}
		},
		"text": func(line gemini.Line) *gemini.LineText {
			switch l := line.(type) {
			case gemini.LineText:
				return &l
			default:
				return nil
			}
		},
		"url": func(ctx *GemtextContext, s string) template.URL {
			u, err := url.Parse(s)
			if err != nil {
				return template.URL("error")
			}
			u = ctx.URL.ResolveReference(u)

			if u.Scheme == "" || u.Scheme == "gemini" {
				if u.Host != ctx.Root.Host {
					u.Path = fmt.Sprintf("/x/%s%s", u.Host, u.Path)
				}
				u.Scheme = ""
				u.Host = ""
			}
			return template.URL(u.String())
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"safeURL": func(s string) template.URL {
			fmt.Println(s)
			f := strings.Replace(s, "localhost", "karelbilek.com", 1)
			fmt.Println(f)

			return template.URL(f)
		},
	}).
	Parse(`<!doctype html>
<html amp lang="en">
<head>
<meta charset="utf-8">
<title>{{.Title}}</title>
<meta name="viewport" content="width=device-width, initial-scale=1" />

<style>
body {
	max-width: 920px;
	margin: 0 auto !important;
	padding: 1rem 2rem;
	font-family: 'Helvetica', 'Arial', sans-serif;
}

</style>
</head>
<body>
<article{{if .Lang}} lang="{{.Lang}}"{{end}}>
	{{ $ctx := . -}}
	{{- $isList := false -}}
	{{- range .Lines -}}
	{{- if and $isList (not (. | li)) }}
	</ul>
	{{- $isList = false -}}
	{{- end -}}

	{{- with . | heading }}
	{{- $isList = false -}}
	<h{{.Level}} id="{{.Anchor}}">{{.Text}}</h{{.Level}}>
	{{- end -}}

	{{- with . | link }}
	{{- $isList = false -}}
	<p>
	<a
		href="{{.URL | url $ctx}}"
	>{{if .Name}}{{.Name}}{{else}}{{.URL}}{{end}}</a>
	{{- end -}}

	{{- with . | quote }}
	{{- $isList = false -}}
	<blockquote>
		{{slice .String 1}}
	</blockquote>
	{{- end -}}

	{{- with . | pre_toggle_on $ctx }}
	<div aria-label="{{slice .String 3}}">
		<pre aria-hidden="true" alt="{{slice .String 3}}">
	{{- $isList = false -}}
	{{- end -}}
	{{- with . | pre -}}
	{{- $isList = false -}}
	{{.}}
{{ end -}}
	{{- with . | pre_toggle_off $ctx -}}
	{{- $isList = false -}}
		</pre>
	</div>
	{{- end -}}

	{{- with . | text }}
	{{- $isList = false }}
	<p>{{.}}
	{{- end -}}

	{{- with . | li }}
	{{- if not $isList }}
	<ul>
	{{- end -}}

	{{- $isList = true }}
		<li>{{slice .String 1}}</li>
	{{- end -}}

	{{- end }}
	{{- if $isList }}
	</ul>
	{{- end }}
</article>
	<hr>
		<a href="https://github.com/karelbilek/website">github</a> | <a href="gemini://karelbilek.com">gemini</a>
</body>
</html>
`))

//

const defaultCSS = ``

type GemtextContext struct {
	CSS         string
	ExternalCSS bool
	External    bool
	Lines       []gemini.Line
	Pre         int
	Resp        *gemini.Response
	Title       string
	Lang        string
	URL         *url.URL
	Root        *url.URL
}

type InputContext struct {
	CSS         string
	ExternalCSS bool
	Prompt      string
	Secret      bool
	URL         *url.URL
}

type GemtextHeading struct {
	Level  int
	Text   string
	Anchor string
}

func createAnchor(heading string) string {
	var anchor strings.Builder
	prev := '-'
	for _, c := range heading {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			anchor.WriteRune(unicode.ToLower(c))
			prev = c
		} else if (unicode.IsSpace(c) || c == '-') && prev != '-' {
			anchor.WriteRune('-')
			prev = '-'
		}
	}
	return strings.ToLower(anchor.String())
}

func proxyGeminiExternal(req gemini.Request,
	w http.ResponseWriter) {

	w.Header().Add("Content-Type", "text/html")

	url := req.URL.String()
	text := `%s<br><br>You've followed a link to another Gemini server!  The <tt>karelbilek.com</tt> server mirrors its own Gemini content.  But if you want to explore the rest of Geminispace, you'll have to use a proper Gemini client that speaks the protocol natively.  Don't worry, <a href="source:https://gemini.circumlunar.space/clients.html">there are lots of clients to choose from</a> for all major platforms, including Android and iOS.`

	text = fmt.Sprintf(text, url)

	w.Write(([]byte)(text))
}

func proxyGemini(req gemini.Request, external bool, root *url.URL,
	w http.ResponseWriter, r *http.Request, css string, externalCSS bool) {

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	client := gemini.Client{}
	resp, err := client.Do(ctx, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Gateway error: %v", err)
		return
	}
	defer resp.Body.Close()

	switch resp.Status {
	case 10, 11:
		w.WriteHeader(http.StatusNotFound)
		// I don't have inputs on my site so who cares
		fmt.Fprintf(w, "Strange times,")
		return
	case 20:
		break // OK
	case 30, 31:
		to, err := url.Parse(resp.Meta)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "Gateway error: bad redirect: %v", err)
		}
		next := req.URL.ResolveReference(to)
		if next.Scheme != "gemini" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "This page is redirecting you to %s", next)
			return
		}
		if external {
			panic("no")
			next.Path = fmt.Sprintf("/x/%s/%s", next.Host, next.Path)
		}
		next.Host = r.URL.Host
		next.Scheme = r.URL.Scheme
		w.Header().Add("Location", next.String())
		w.WriteHeader(http.StatusFound)
		fmt.Fprintf(w, "Redirecting to %s", next)
		return
	case 40, 41, 42, 43, 44:
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "The remote server returned %d: %s", resp.Status, resp.Meta)
		return
	case 50, 51:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "The remote server returned %d: %s", resp.Status, resp.Meta)
		return
	case 52, 53, 59:
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "The remote server returned %d: %s", resp.Status, resp.Meta)
		return
	default:
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "Proxy does not understand Gemini response status %d", resp.Status)
		return
	}

	m, params, err := mime.ParseMediaType(resp.Meta)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Gateway error: %d %s: %v", resp.Status, resp.Meta, err)
		return
	}

	if m != "text/gemini" {
		w.Header().Add("Content-Type", resp.Meta)
		io.Copy(w, resp.Body)
		return
	}

	if charset, ok := params["charset"]; ok {
		charset = strings.ToLower(charset)
		if charset != "utf-8" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "Unsupported charset: %s", charset)
			return
		}
	}

	lang := params["lang"]

	w.Header().Add("Content-Type", "text/html")
	gemctx := &GemtextContext{
		CSS:         css,
		ExternalCSS: externalCSS,
		External:    external,
		Resp:        resp,
		Title:       "Karel Bilek",
		Lang:        lang,
		URL:         req.URL,
		Root:        root,
	}

	var title bool
	gemini.ParseLines(resp.Body, func(line gemini.Line) {
		gemctx.Lines = append(gemctx.Lines, line)
		if !title {
			if h, ok := line.(gemini.LineHeading1); ok {
				gemctx.Title = string(h)
				title = true
			}
		}
	})

	err = gemtextPage.Execute(w, gemctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}
}

/*
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
*/

type HandleFunc func(http.ResponseWriter, *http.Request)

func (h HandleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

// hack for legacy comics page, TODO fix later
//
//go:embed chronocomics/dist/bundle.js
var comicsBundlejs []byte

// hack for legacy comics page, TODO fix later
//
//go:embed chronocomics/dist/index.html
var comicsIndex []byte

//go:embed static/*
var f embed.FS

func mainProxy() {
	var (
		css      string = defaultCSS
		external bool   = false
	)

	root, err := url.Parse("gemini://localhost")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("POST /echo-post", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "error on parsing form: %+v", err)
			return
		}
		js, _ := json.MarshalIndent(r.Form, "", "   ")
		w.Write(js)
	})

	http.Handle("GET /static/", http.FileServerFS(f))
	http.Handle("GET data.karelbilek.com/", http.FileServerFS(f))

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Printf("%s %s", r.Method, r.URL.Path)
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("404 Not found"))
			return
		}

		// hack for old comics site
		if strings.Contains(r.URL.Path, "chronocomics") {
			if strings.Contains(r.URL.Path, "bundle.js") {
				w.Write(comicsBundlejs)
			} else {
				w.Write(comicsIndex)
			}
			return
		}

		if r.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not found"))
			return
		}

		if r.URL.Path == "/visits.txt" {
			visits, err := logs.LatestLogsText()
			if err != nil {
				log.Printf("cannot get visits: %+v", err)
				w.Write([]byte("cannot get visits"))
				return
			}
			w.Write([]byte(visits))
			return
		}

		err := logs.Mark(r.URL.Path, false)
		if err != nil {
			log.Printf("cannot log to sql: %+v", err)
		}

		req := gemini.Request{}
		req.URL = &url.URL{}
		req.URL.Scheme = root.Scheme
		req.URL.Host = root.Host
		// as this is public code, someone can "hack" this, but all that it will do is to not display
		// in the SQL logs, which is... whatever
		req.URL.Path = r.URL.Path + "_proxied"
		req.URL.RawQuery = r.URL.RawQuery
		proxyGemini(req, false, root, w, r, css, external)
	}))

	var bind string = ":8080"
	log.Printf("HTTP server listening on %s", bind)
	log.Fatal(http.ListenAndServe(bind, nil))

}
