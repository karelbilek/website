package logs

import (
	"database/sql"
	"errors"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

var globalDB *sql.DB

func init() {
	path := "/app/data/data.db"
	if _, err := os.Stat("/app/data"); errors.Is(err, os.ErrNotExist) {
		path = "data/data.db"
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visits (url STRING NOT NULL, gemini INTEGER, time INTEGER)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS visits_time ON visits(time)`)
	if err != nil {
		panic(err)
	}

	globalDB = db
}

func Mark(url string, gemini bool) error {
	gem := 0
	if gemini {
		gem = 1
	}
	_, err := globalDB.Exec(`INSERT INTO visits(url, gemini, time) VALUES(?, ?, unixepoch('now'))`, url, gem)
	return err
}

type Log struct {
	URL    string
	Gemini bool
	Time   time.Time
}

func LatestLogsText() (string, error) {
	logs, err := LatestLogs()
	if err != nil {
		return "", err
	}

	res := ""
	for _, l := range logs {
		res += l.Time.Format(time.RFC1123) + " "
		res += l.URL
		if l.Gemini {
			res += " gemini"
		} else {
			res += " http"
		}
		res += "\n"
	}
	return res, nil
}

func LatestLogs() ([]Log, error) {
	rows, err := globalDB.Query(`SELECT url, gemini, time FROM visits ORDER BY time ASC LIMIT 50`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res := []Log{}
	type SQLLog struct {
		URL    string
		Gemini int
		Time   int
	}
	for rows.Next() {
		var l SQLLog
		if err := rows.Scan(&l.URL, &l.Gemini, &l.Time); err != nil {
			return nil, err
		}
		var ll Log
		ll.URL = l.URL
		ll.Gemini = l.Gemini == 1
		ll.Time = time.Unix(int64(l.Time), 0)
		res = append(res, ll)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
