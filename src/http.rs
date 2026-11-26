//! HTTP side: proxies the gemtext site as HTML, plus static files and small hacks.

use crate::site::{self, Page};
use crate::visits::Visits;
use axum::extract::{Form, Request, State};
use axum::http::{StatusCode, Uri, header};
use axum::middleware::{self, Next};
use axum::response::{Html, IntoResponse, Redirect, Response};
use axum::routing::{get, post};
use axum::{Router, ServiceExt};
use std::collections::BTreeMap;
use std::sync::Arc;
use tower::Layer;
use tower_http::services::ServeDir;

pub async fn serve(visits: Arc<Visits>) -> std::io::Result<()> {
    let router = Router::new()
        .route("/echo-post", post(echo_post))
        .nest_service("/static", ServeDir::new("static").redirect_path_prefix("/static"))
        .fallback(get(page))
        .with_state(visits);
    // runs before routing, so it can rewrite the path
    let app = middleware::from_fn(data_host).layer(router);

    let port = std::env::var("PORT").unwrap_or_else(|_| "8080".into());
    let listener = tokio::net::TcpListener::bind(format!("0.0.0.0:{port}")).await?;
    tracing::info!("http server listening on :{port}");
    axum::serve(listener, ServiceExt::<Request>::into_make_service(app)).await
}

/// data.karelbilek.com serves `static/` at its root.
async fn data_host(mut req: Request, next: Next) -> Response {
    let is_data = req
        .headers()
        .get(header::HOST)
        .and_then(|h| h.to_str().ok())
        .is_some_and(|h| h.starts_with("data.karelbilek.com"));
    if is_data {
        let pq = req.uri().path_and_query().map(|p| p.as_str()).unwrap_or("/");
        if let Ok(uri) = format!("/static{pq}").parse::<Uri>() {
            *req.uri_mut() = uri;
        }
    }
    next.run(req).await
}

async fn echo_post(Form(fields): Form<Vec<(String, String)>>) -> String {
    let mut m: BTreeMap<String, Vec<String>> = BTreeMap::new();
    for (k, v) in fields {
        m.entry(k).or_default().push(v);
    }
    serde_json::to_string_pretty(&m).unwrap_or_default()
}

async fn page(State(visits): State<Arc<Visits>>, uri: Uri) -> Response {
    let path = uri.path();
    tracing::info!("GET {path}");

    // hack for old comics site
    if path.contains("chronocomics") {
        let (file, mime) = if path.contains("bundle.js") {
            ("chronocomics/dist/bundle.js", "application/javascript")
        } else {
            ("chronocomics/dist/index.html", "text/html")
        };
        return match std::fs::read(file) {
            Ok(body) => ([(header::CONTENT_TYPE, mime)], body).into_response(),
            Err(_) => StatusCode::NOT_FOUND.into_response(),
        };
    }
    if path == "/favicon.ico" {
        return (StatusCode::NOT_FOUND, "404 Not found").into_response();
    }
    if path == "/visits.txt" {
        return match visits.latest_text() {
            Ok(t) => t.into_response(),
            Err(e) => {
                tracing::error!("cannot get visits: {e}");
                "cannot get visits".into_response()
            }
        };
    }

    visits.mark(path, false);
    match site::get(path, "") {
        Page::Ok { mime, body } if mime == site::GEMINI_MIME => {
            Html(crate::gemtext::to_html(&String::from_utf8_lossy(&body))).into_response()
        }
        Page::Ok { mime, body } => ([(header::CONTENT_TYPE, mime)], body).into_response(),
        Page::Redirect(to) => Redirect::temporary(&to).into_response(),
        Page::NotFound => (StatusCode::NOT_FOUND, "oopsie woopsie!! UwU").into_response(),
    }
}
