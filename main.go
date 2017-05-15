package main

import (
	"database/sql"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	sqlite "github.com/mattn/go-sqlite3"
)

const (
	pixelRaw   = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
	pixelRoute = "/orkka.gif"
	jsRoute    = "/a.js"
)

type config struct {
	Port                  string `env:"PORT"             envDefault:"8080"`
	DBFile                string `env:"DB_FILE"          envDefault:"events.db"`
	JSFile                string `env:"JS_FILE"          envDefault:"analytics.js"`
	EventQueueSize        int    `env:"MAX_CONNECTIONS"  envDefault:"4096"`
	WriteQueueDefaultSize int    `env:"WRITE_QUEUE_SIZE" envDefault:"1024"`
	WriteFrequencyMillis  int    `env:"WRITE_FREQUENCY"  envDefault:"100"`
}

func main() {
	var config config
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	eventQueue := make(chan *pageEvent, config.EventQueueSize)

	go processEventWriteQueue(&config, eventQueue)
	http.HandleFunc(pixelRoute, makeHandlePixel(eventQueue))
	http.HandleFunc(jsRoute, makeHandleScript(config.JSFile))

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func makeHandleScript(file string) http.HandlerFunc {
	javascript, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Failed to read javascript file: %v", err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		out.Header().Set("Content-Type", "text/javascript")
		out.Header().Set("Cache-Control", "public, max-age=86400")
		out.WriteHeader(http.StatusOK)
		out.Write(javascript)
	}
}

func makeHandlePixel(eventQueue chan *pageEvent) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		log.Fatalf("failed to encode pixel: %v", err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		eventQueue <- pageEventFromRequest(req)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func processEventWriteQueue(config *config, eventQueue chan *pageEvent) {
	db := initDB(config)
	defer db.Close()

	writeFrequency := time.Duration(time.Duration(config.WriteFrequencyMillis) * time.Millisecond)
	writeTimer := time.NewTimer(writeFrequency)

	writeQueue := make([]*pageEvent, 0, config.WriteQueueDefaultSize)

	for {
		select {
		case ev := <-eventQueue:
			writeQueue = append(writeQueue, ev)

		case <-writeTimer.C:
			if len(writeQueue) > 0 {
				writeTimer.Stop()
				writeEvents(db, writeQueue)
				writeQueue = writeQueue[:0]
			}
			writeTimer.Reset(writeFrequency)
		}
	}
}

func writeEvents(db *sql.DB, events []*pageEvent) {
	var ev *pageEvent
	i := -1
	db.Exec("BEGIN TRANSACTION;")
	for i, ev = range events {
		writePageEventToDB(db, ev)
	}
	db.Exec("END TRANSACTION;")

	log.Println("wrote", i+1, "records.")
}

func initDB(config *config) *sql.DB {
	sql.Register("sqlite3_custom", &sqlite.SQLiteDriver{
		ConnectHook: registerSQLiteExtensions,
	})

	db, err := sql.Open("sqlite3_custom", config.DBFile)

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("database is null")
	}

	if _, err = db.Exec(sqlCreatePageEventTable); err != nil {
		log.Fatal(err)
	}

	return db
}

const (
	pageEventSchemaVersion  = 1
	sqlCreatePageEventTable = `
		CREATE TABLE IF NOT EXISTS events (
			script_version   integer,
			datetime         datetime NOT NULL,
			server           text,
			remote_addr      text,
			user_agent       text,
			request_referrer text,
			title            text,
			referrer         text,
			url              text,
			event_type       text,
			session_token    text,
			user_token       text
		);

		CREATE INDEX IF NOT EXISTS idx1 ON events(url, referrer);
		CREATE INDEX IF NOT EXISTS idx2 ON events(user_token, session_token, event_type);

		PRAGMA journal_mode = WAL;
	`
	sqlInsertPageEvent = `
		INSERT INTO events(
			script_version,
			datetime,
			server,
			remote_addr,
			user_agent,
			request_referrer,
			title,
			referrer,
			url,
			event_type,
			session_token,
			user_token
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
)

type pageEvent struct {
	ScriptVersion   string    `json:"script_version"`
	Time            time.Time `json:"datetime"`
	Host            string    `json:"server"`
	RemoteAddr      string    `json:"remote_addr"`
	UserAgent       string    `json:"user_agent"`
	RequestReferrer string    `json:"request_referrer"`
	Title           string    `json:"title"`
	PageReferrer    string    `json:"referrer"`
	URL             string    `json:"url"`
	EventType       string    `json:"event_type"`
	SessionToken    string    `json:"session_token"`
	UserToken       string    `json:"user_token"`
	ViewIndex       string    `json:"view_index"`
}

func pageEventFromRequest(req *http.Request) *pageEvent {
	params := req.URL.Query()

	return &pageEvent{
		Host:            req.Host,
		RemoteAddr:      req.RemoteAddr,
		UserAgent:       req.UserAgent(),
		RequestReferrer: req.Referer(),
		Time:            time.Now(),
		Title:           params.Get("t"),
		PageReferrer:    params.Get("ref"),
		URL:             params.Get("url"),
		ScriptVersion:   params.Get("ver"),
		EventType:       params.Get("evt"),
		SessionToken:    params.Get("st"),
		UserToken:       params.Get("ut"),
		ViewIndex:       params.Get("vwix"),
	}
}

func writePageEventToDB(db *sql.DB, ev *pageEvent) {
	stmt, err := db.Prepare(sqlInsertPageEvent)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		ev.ScriptVersion,
		ev.Time,
		ev.Host,
		ev.RemoteAddr,
		ev.UserAgent,
		ev.RequestReferrer,
		ev.Title,
		ev.PageReferrer,
		ev.URL,
		ev.EventType,
		ev.SessionToken,
		ev.UserToken,
	)

	if err != nil {
		log.Fatal(err)
	}
}
