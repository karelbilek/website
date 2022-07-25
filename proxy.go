package main

import (
	"context"
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

	"git.sr.ht/~adnano/go-gemini"
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
<html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1" />
{{- if .CSS }}
{{- if .ExternalCSS }}
<link rel="stylesheet" type="text/css" href="{{.CSS | safeCSS}}">
{{- else }}
<style>
{{.CSS | safeCSS}}
</style>
{{- end }}
{{- end }}
<title>{{.Title}}</title>
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
		Proxied from <a href="{{.URL.String | safeURL}}">{{.URL.String | safeURL}}</a>
		by modified <a href="https://sr.ht/~sircmpwn/kineto">kineto</a>.

`))

var inputPage = template.Must(template.
	New("input").
	Funcs(template.FuncMap{
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	}).
	Parse(`<!doctype html>
<html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1" />
{{- if .CSS }}
{{- if .ExternalCSS }}
<link rel="stylesheet" type="text/css" href="{{.CSS | safeCSS}}">
{{- else }}
<style>
{{.CSS | safeCSS}}
</style>
{{- end }}
{{- end }}
<title>{{.Prompt}}</title>
<form method="POST">
	<label for="input">{{.Prompt}}</label>
	{{ if .Secret }}
	<input type="password" id="input" name="q" />
	{{ else }}
	<input type="text" id="input" name="q" />
	{{ end }}
</form>
`))

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
		w.Header().Add("Content-Type", "text/html")
		err = inputPage.Execute(w, &InputContext{
			CSS:         css,
			ExternalCSS: externalCSS,
			Prompt:      resp.Meta,
			Secret:      resp.Status == 11,
			URL:         req.URL,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
		}
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
		Title:       req.URL.Host + " " + req.URL.Path,
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

func mainProxy() {
	var (
		css      string = defaultCSS
		external bool   = false
	)

	root, err := url.Parse("gemini://localhost")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//if r.Method == "POST" {
		//	r.ParseForm()
		//	if q, ok := r.Form["q"]; !ok {
		//		w.WriteHeader(http.StatusBadRequest)
		//		w.Write([]byte("Bad request"))
		//	} else {
		//		w.Header().Add("Location", "?"+q[0])
		//		w.WriteHeader(http.StatusFound)
		//		w.Write([]byte("Redirecting"))
		//	}
		//	return
		//}

		log.Printf("%s %s", r.Method, r.URL.Path)
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("404 Not found"))
			return
		}

		if r.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not found"))
			return
		}
		//
		//if r.URL.Path == "/robots.txt" {
		//	w.WriteHeader(http.StatusOK)
		//	w.Write([]byte("User-agent: *\nDisallow: /\n"))
		//	return
		//}

		req := gemini.Request{}
		req.URL = &url.URL{}
		req.URL.Scheme = root.Scheme
		req.URL.Host = root.Host
		req.URL.Path = r.URL.Path
		req.URL.RawQuery = r.URL.RawQuery
		proxyGemini(req, false, root, w, r, css, external)
	}))

	http.Handle("/x/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			if q, ok := r.Form["q"]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad request"))
			} else {
				w.Header().Add("Location", "?"+q[0])
				w.WriteHeader(http.StatusFound)
				w.Write([]byte("Redirecting"))
			}
			return
		}

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("404 Not found"))
			return
		}

		path := strings.SplitN(r.URL.Path, "/", 4)
		if len(path) != 4 {
			path = append(path, "")
		}
		req := gemini.Request{}
		req.URL = &url.URL{}
		req.URL.Scheme = "gemini"
		req.URL.Host = path[2]
		req.URL.Path = "/" + path[3]
		req.URL.RawQuery = r.URL.RawQuery
		log.Printf("%s (external) %s%s", r.Method, r.URL.Host, r.URL.Path)
		proxyGeminiExternal(req, w)
	}))

	//if _, err := os.Stat("/etc/letsencrypt/live/karelbilek.com/fullchain.pem"); errors.Is(err, os.ErrNotExist) {
	var bind string = ":80"
	log.Printf("HTTP server listening on %s", bind)
	log.Fatal(http.ListenAndServe(bind, nil))
	// } else {
	// 	go func() {
	// 		handler := HandleFunc(func(writer http.ResponseWriter, request *http.Request) {
	// 			http.Redirect(writer, request, "https://karelbilek.com", 301)
	// 		})
	// 		var bind string = ":80"
	// 		log.Printf("redirect server listening on %s", bind)
	// 		log.Fatal(http.ListenAndServe(bind, handler))
	// 	}()

	// 	var bind string = ":443"
	// 	log.Printf("HTTPS server listening on %s", bind)
	// 	log.Fatal(http.ListenAndServeTLS(bind, "/etc/letsencrypt/live/karelbilek.com/fullchain.pem", "/etc/letsencrypt/live/karelbilek.com/privkey.pem", nil))
	// }
}
