use rusqlite::{Connection, params};
use std::sync::Mutex;

/// Visit log in SQLite. Same file and schema as the old Go version.
pub struct Visits(Mutex<Connection>);

impl Visits {
    pub fn open() -> rusqlite::Result<Self> {
        let path = if std::path::Path::new("/app/data").is_dir() {
            "/app/data/data.db"
        } else {
            "data/data.db"
        };
        let db = Connection::open(path)?;
        db.execute_batch(
            "CREATE TABLE IF NOT EXISTS visits (url STRING NOT NULL, gemini INTEGER, time INTEGER);
             CREATE INDEX IF NOT EXISTS visits_time ON visits(time);",
        )?;
        Ok(Self(Mutex::new(db)))
    }

    pub fn mark(&self, url: &str, gemini: bool) {
        let res = self.0.lock().unwrap().execute(
            "INSERT INTO visits(url, gemini, time) VALUES(?1, ?2, unixepoch('now'))",
            params![url, gemini as i32],
        );
        if let Err(e) = res {
            tracing::error!("cannot log to sql: {e}");
        }
    }

    /// Last 50 visits, one per line.
    pub fn latest_text(&self) -> rusqlite::Result<String> {
        let db = self.0.lock().unwrap();
        let mut stmt = db.prepare(
            "SELECT datetime(time, 'unixepoch'), url, gemini FROM visits ORDER BY time DESC LIMIT 50",
        )?;
        let rows = stmt.query_map([], |r| {
            let proto = if r.get::<_, i32>(2)? == 1 { "gemini" } else { "http" };
            Ok(format!("{} {} {}\n", r.get::<_, String>(0)?, r.get::<_, String>(1)?, proto))
        })?;
        rows.collect()
    }
}
