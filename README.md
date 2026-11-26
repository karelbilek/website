My current webpage.

It gets hosted on fly.io for free (!).

It is now running at https://karelbilek.com

Slop-rewritten by Claude into Rust (axum + tokio-rustls + rusqlite). `cargo run` locally; needs `gemini-cert.pem` and `gemini-key.rsa` in cwd. `PORT` env sets HTTP port (default 8080), gemini is always 1965.

Functionality of this hacked-together code:

* it gets hosted both as HTTP and Gemini
  * Gemini is a simple Gopher-like protocol that nobody uses
  * see gemini://karelbilek.com with LaGrange browser
  * 90% of complexity in this code is that it serves both gemini and http
  * HTTP side renders the same gemtext files to HTML directly; original idea from https://git.sr.ht/~sircmpwn/kineto
  * I keep using the gemini stuff even when it has 0 views; I like how it forces me to make the webpage ridiculously simple
* it auto-deploys on fly.io using github workflows and a plain Dockerfile
* it saves visit data to SQLite that is also on fly.io and displays them on https://karelbilek.com/visits.txt
* it works both locally with `cargo run` and with deployments to fly.io, including the SQLite
* chronocomics has index of supermegamonkey comics archive, and some experiments with react/redux
* chronocomics/extractor is a standalone go module (the scraper)
* the actual content is in public/, in raw gemini format

It is licensed GPLv3; originally because I used kineto directly, now I slop-rewrote most of it, but eh who cares