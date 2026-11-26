mod gemini;
mod gemtext;
mod http;
mod site;
mod visits;

use std::sync::Arc;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()),
        )
        .init();

    let visits = Arc::new(visits::Visits::open().expect("open sqlite"));

    let v = visits.clone();
    tokio::spawn(async move {
        if let Err(e) = gemini::serve(v).await {
            tracing::error!("gemini server died: {e}");
            std::process::exit(1);
        }
    });

    if let Err(e) = http::serve(visits).await {
        tracing::error!("http server died: {e}");
        std::process::exit(1);
    }
}
