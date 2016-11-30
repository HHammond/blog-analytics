package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/mattn/go-sqlite3"
)

const (
	pixelRaw   = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
	pixelRoute = "/a.gif"
)

type config struct {
	Port                  string `env:"PORT"             envDefault:"8080"`
	DBFile                string `env:"DB_FILE"          envDefault:"events.db"`
	EventQueueSize        int    `env:"MAX_CONNECTIONS"  envDefault:"1000"`
	WriteQueueDefaultSize int    `env:"WRITE_QUEUE_SIZE" envDefault:"10000"`
	WriteFrequencyMillis  int    `env:"WRITE_FREQUENCY"  envDefault:"10"`
}

func main() {
	var config config
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	db := initDB(&config)
	defer db.Close()

	eventQueue := make(chan *pageEvent, config.EventQueueSize)

	go handleWriteEventQueue(&config, db, eventQueue)
	http.HandleFunc(pixelRoute, makeHandlePixel(eventQueue))

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
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

func handleWriteEventQueue(config *config, db *sql.DB, eventQueue chan *pageEvent) {
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

				db.Exec("BEGIN TRANSACTION;")
				for _, ev := range writeQueue {
					writePageEventToDB(db, ev)
				}
				db.Exec("END TRANSACTION;")

				log.Println("wrote", len(writeQueue), "records.")
				writeQueue = writeQueue[:0]
			}
			writeTimer.Reset(writeFrequency)
		}
	}
}

func initDB(config *config) *sql.DB {
	db, err := sql.Open("sqlite3", config.DBFile)

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("database is null")
	}

	_, err = db.Exec(sqlCreatePageEventTable)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

const (
	pageEventSchemaVersion  = 1
	sqlCreatePageEventTable = `
		CREATE TABLE IF NOT EXISTS events (
			schema_version   integer,
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

		PRAGMA journal_mode = MEMORY;
	`
	sqlInsertPageEvent = `
		INSERT INTO events(
			schema_version,
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
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
	}
}

func writePageEventToDB(db *sql.DB, ev *pageEvent) {
	stmt, err := db.Prepare(sqlInsertPageEvent)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		pageEventSchemaVersion,
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
