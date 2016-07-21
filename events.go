//go:generate reform

package main

import (
	"database/sql"
	"net/http"
	"time"
)

const (
	PageEventSchemaVersion  = 1
	SQLPageEventCreateTable = `
		CREATE TABLE IF NOT EXISTS page_events (
			schema_version   integer,
			script_version   integer,
			datetime		 datetime NOT NULL,
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

		PRAGMA journal_mode = WAL;
	`
	SQLPageEventInsert = `
		INSERT INTO page_events(
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

//reform:page_events
type PageEvent struct {
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

func PageEventFromRequest(req *http.Request) *PageEvent {
	params := req.URL.Query()
	return &PageEvent{
		Host:            req.Host,
		RemoteAddr:      req.RemoteAddr,
		UserAgent:       req.UserAgent(),
		RequestReferrer: req.Referer(),
		Time:            time.Now(),
		Title:           params.Get("title"),
		PageReferrer:    params.Get("referrer"),
		URL:             params.Get("url"),
		ScriptVersion:   params.Get("version"),
		EventType:       params.Get("event_type"),
		SessionToken:    params.Get("session_token"),
		UserToken:       params.Get("user_token"),
	}
}

func (ev PageEvent) InsertIntoDB(db *sql.DB) {
	stmt, err := db.Prepare(SQLPageEventInsert)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		PageEventSchemaVersion,
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
		panic(err)
	}
}
