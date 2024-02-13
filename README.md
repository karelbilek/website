My current webpage.

It gets hosted on fly.io for free (!).

It is now running at https://karelbilek.com

Functionality of this hacked-together code:

* it gets hosted both as HTTP and Gemini
  * Gemini is a simple Gopher-like protocol that nobody uses
  * see gemini://karelbilek.com with LaGrange browser
  * 90% of complexity in this code is that it serves both gemini and http
  * it is using https://git.sr.ht/~sircmpwn/kineto and https://github.com/n0x1m/gmifs code, but very heavily modified. I will probably simplify it further if I ever have time (not).
  * I keep using the gemini stuff even when it has 0 views; I like how it forces me to make the webpage ridiculously simple
* it auto-deployes on fly.io using github workflows
* it saves visit data to SQLite that is also on fly.io and displays them on https://karelbilek.com/visits.txt
* it works both locally with `go run .` and with deployments to fly.io, including the SQLite
* choronocomics has index of supermegamonkey comics archive, and some experiments with react/redux
* the actual content is in public/, in raw gemini format

It is licensed GPLv3, mainly because kineto is