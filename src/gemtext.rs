//! Minimal gemtext parser and HTML renderer.

use html_escape::{encode_double_quoted_attribute as attr, encode_text as text};
use std::fmt::Write;

#[derive(Debug, PartialEq)]
pub enum Line<'a> {
    Text(&'a str),
    Link { url: &'a str, name: &'a str },
    Heading(u8, &'a str),
    ListItem(&'a str),
    Quote(&'a str),
    /// Opens or closes a preformatted block; payload is the alt text.
    PreToggle(&'a str),
    Pre(&'a str),
}

pub fn parse(src: &str) -> Vec<Line<'_>> {
    let mut pre = false;
    src.lines()
        .map(|l| l.trim_end_matches('\r'))
        .map(|l| {
            if let Some(alt) = l.strip_prefix("```") {
                pre = !pre;
                Line::PreToggle(alt.trim())
            } else if pre {
                Line::Pre(l)
            } else if let Some(rest) = l.strip_prefix("=>") {
                let rest = rest.trim();
                let (url, name) = rest.split_once(char::is_whitespace).unwrap_or((rest, ""));
                Line::Link { url, name: name.trim() }
            } else if let Some(h) = l.strip_prefix("###") {
                Line::Heading(3, h.trim())
            } else if let Some(h) = l.strip_prefix("##") {
                Line::Heading(2, h.trim())
            } else if let Some(h) = l.strip_prefix('#') {
                Line::Heading(1, h.trim())
            } else if let Some(li) = l.strip_prefix("* ") {
                Line::ListItem(li.trim())
            } else if let Some(q) = l.strip_prefix('>') {
                Line::Quote(q.trim())
            } else {
                Line::Text(l)
            }
        })
        .collect()
}

fn is_image(url: &str) -> bool {
    let path = url.split(['?', '#']).next().unwrap_or("");
    let ext = path.rsplit('.').next().unwrap_or("").to_ascii_lowercase();
    matches!(ext.as_str(), "png" | "jpg" | "jpeg" | "gif" | "webp" | "avif" | "svg")
}

fn anchor(heading: &str) -> String {
    let mut out = String::new();
    for c in heading.chars() {
        if c.is_alphanumeric() {
            out.extend(c.to_lowercase());
        } else if (c.is_whitespace() || c == '-') && !out.is_empty() && !out.ends_with('-') {
            out.push('-');
        }
    }
    out
}

/// Renders a full HTML page from gemtext.
pub fn to_html(src: &str) -> String {
    let lines = parse(src);
    let title = lines
        .iter()
        .find_map(|l| match l {
            Line::Heading(1, t) => Some(*t),
            _ => None,
        })
        .unwrap_or("Karel Bilek");

    let mut b = String::new();
    let mut in_list = false;
    let mut in_pre = false;
    for line in &lines {
        if in_list && !matches!(line, Line::ListItem(_)) {
            b.push_str("</ul>\n");
            in_list = false;
        }
        match line {
            Line::Text("") => {}
            Line::Text(t) => writeln!(b, "<p>{}</p>", text(t)).unwrap(),
            Line::Link { url, name } => {
                let name = if name.is_empty() { url } else { name };
                if is_image(url) {
                    // closed <details> plus lazy img: nothing is fetched until clicked
                    writeln!(
                        b,
                        "<details><summary>\u{1f5bc}\u{fe0f} {}</summary><img src=\"{}\" alt=\"{}\" loading=\"lazy\" style=\"max-width:100%\"></details>",
                        text(name), attr(url), attr(name)
                    )
                    .unwrap()
                } else {
                    writeln!(b, "<p><a href=\"{}\">{}</a></p>", attr(url), text(name)).unwrap()
                }
            }
            Line::Heading(n, t) => {
                writeln!(b, "<h{n} id=\"{}\">{}</h{n}>", anchor(t), text(t)).unwrap()
            }
            Line::ListItem(t) => {
                if !in_list {
                    b.push_str("<ul>\n");
                    in_list = true;
                }
                writeln!(b, "<li>{}</li>", text(t)).unwrap()
            }
            Line::Quote(t) => writeln!(b, "<blockquote>{}</blockquote>", text(t)).unwrap(),
            Line::PreToggle(alt) => {
                in_pre = !in_pre;
                if in_pre {
                    writeln!(b, "<pre aria-label=\"{}\">", attr(alt)).unwrap()
                } else {
                    b.push_str("</pre>\n")
                }
            }
            Line::Pre(t) => writeln!(b, "{}", text(t)).unwrap(),
        }
    }
    if in_list {
        b.push_str("</ul>\n");
    }
    if in_pre {
        b.push_str("</pre>\n");
    }

    format!(
        r#"<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{}</title>
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
body {{
	max-width: 920px;
	margin: 0 auto !important;
	padding: 1rem 2rem;
	font-family: 'Helvetica', 'Arial', sans-serif;
}}
summary {{
	cursor: pointer;
	color: -webkit-link;
	color: LinkText;
	text-decoration: underline;
}}
</style>
</head>
<body>
<article>
{}</article>
<hr>
<a href="https://github.com/karelbilek/website">github</a> | <a href="gemini://karelbilek.com">gemini</a>
</body>
</html>
"#,
        text(title),
        b
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_and_render() {
        let src = "# Hi there\n=> foo.gmi Foo\n=> bar\n* a\n* b\ntext <b>\n```alt\n# not heading\n```\n> q";
        let lines = parse(src);
        assert_eq!(lines[0], Line::Heading(1, "Hi there"));
        assert_eq!(lines[1], Line::Link { url: "foo.gmi", name: "Foo" });
        assert_eq!(lines[2], Line::Link { url: "bar", name: "" });
        assert_eq!(lines[7], Line::Pre("# not heading"));
        let html = to_html(src);
        assert!(html.contains("<title>Hi there</title>"));
        assert!(html.contains("<h1 id=\"hi-there\">Hi there</h1>"));
        assert!(html.contains("<a href=\"bar\">bar</a>"));
        assert!(html.contains("<ul>\n<li>a</li>\n<li>b</li>\n</ul>"));
        assert!(html.contains("<p>text &lt;b&gt;</p>"));
        assert!(html.contains("<pre aria-label=\"alt\">\n# not heading\n</pre>"));
        assert!(html.contains("<blockquote>q</blockquote>"));
        let img = to_html("=> pie.PNG A pie");
        assert!(img.contains("<summary>\u{1f5bc}\u{fe0f} A pie</summary><img src=\"pie.PNG\" alt=\"A pie\" loading=\"lazy\""));
        assert!(!is_image("foo.gmi") && is_image("x/y.webp?v=1"));
    }
}
