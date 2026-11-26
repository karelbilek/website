//! Tiny gemini protocol server: TLS, one request line, one response.

use crate::site::{self, GMI_NOTE, Page};
use crate::visits::Visits;
use rustls_pki_types::{CertificateDer, PrivateKeyDer, pem::PemObject};
use std::sync::Arc;
use std::time::Duration;
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};
use tokio_rustls::{TlsAcceptor, rustls::ServerConfig, server::TlsStream};

pub async fn serve(visits: Arc<Visits>) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let certs = CertificateDer::pem_file_iter("gemini-cert.pem")?.collect::<Result<Vec<_>, _>>()?;
    let key = PrivateKeyDer::from_pem_file("gemini-key.rsa")?;
    let config = ServerConfig::builder().with_no_client_auth().with_single_cert(certs, key)?;
    let acceptor = TlsAcceptor::from(Arc::new(config));

    let listener = TcpListener::bind("0.0.0.0:1965").await?;
    tracing::info!("gemini server listening on :1965");
    loop {
        let (sock, addr) = listener.accept().await?;
        let acceptor = acceptor.clone();
        let visits = visits.clone();
        tokio::spawn(async move {
            let Ok(mut tls) = acceptor.accept(sock).await else { return };
            let res = tokio::time::timeout(Duration::from_secs(5), handle(&mut tls, &visits)).await;
            match res {
                Ok(Ok(line)) => tracing::info!("gem {addr} {line}"),
                Ok(Err(e)) => tracing::warn!("gem {addr} error: {e}"),
                Err(_) => tracing::warn!("gem {addr} timeout"),
            }
            let _ = tls.shutdown().await;
        });
    }
}

/// Returns a one-line summary for the log.
async fn handle(tls: &mut TlsStream<TcpStream>, visits: &Visits) -> std::io::Result<String> {
    let mut line = String::new();
    BufReader::new(&mut *tls).take(1026).read_line(&mut line).await?;
    let line = line.trim_end_matches(['\r', '\n']);

    let (header, body) = match url::Url::parse(line) {
        Ok(u) if u.scheme() == "gemini" => {
            visits.mark(u.path(), true);
            match site::get(u.path(), GMI_NOTE) {
                Page::Ok { mime, body } => (format!("20 {mime}"), body),
                Page::Redirect(to) => (format!("31 {to}"), vec![]),
                Page::NotFound => ("51 oopsie woopsie!! UwU".to_string(), vec![]),
            }
        }
        Ok(_) => ("53 proxy request refused".to_string(), vec![]),
        Err(_) => ("59 bad request".to_string(), vec![]),
    };
    tls.write_all(format!("{header}\r\n").as_bytes()).await?;
    tls.write_all(&body).await?;
    tls.flush().await?;
    Ok(format!("{line:?} {}", &header[..2]))
}
