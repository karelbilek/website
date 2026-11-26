//! Serves the gemtext site from `public/`. Shared by the gemini and HTTP servers.

use std::path::{Component, Path, PathBuf};

pub const GEMINI_MIME: &str = "text/gemini; charset=utf-8";

/// Shown in place of `<GMI>` to gemini readers only.
// pub const GMI_NOTE: &str = "I am thinking what is the point of running gemini server. If you are reading this on gemini, can you drop me a mail on kb@karelbilek.com ?";
pub const GMI_NOTE: &str = "";

pub enum Page {
    Ok { mime: &'static str, body: Vec<u8> },
    Redirect(String),
    NotFound,
}

/// `gmi_note` replaces `<GMI>` in gemtext bodies.
pub fn get(req_path: &str, gmi_note: &str) -> Page {
    let rel = req_path.trim_matches('/');
    if !Path::new(rel).components().all(|c| matches!(c, Component::Normal(_))) {
        return Page::NotFound;
    }
    let mut file = PathBuf::from("public").join(rel);
    if file.is_dir() {
        if !req_path.ends_with('/') {
            return Page::Redirect(format!("{req_path}/"));
        }
        file.push("index.gmi");
    }
    let Ok(body) = std::fs::read(&file) else {
        return Page::NotFound;
    };
    if file.extension().is_some_and(|e| e == "gmi") {
        let body = String::from_utf8_lossy(&body).replace("<GMI>", gmi_note);
        return Page::Ok { mime: GEMINI_MIME, body: body.into_bytes() };
    }
    match mime_guess::from_path(&file).first_raw() {
        Some(mime) => Page::Ok { mime, body },
        None => Page::NotFound,
    }
}
