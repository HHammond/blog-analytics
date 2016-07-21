//go:generate reform

package main

import (
	"net/http"
	"time"
)

const (
	PageEventSchemaVersion = 1
	CreatePageEventTable   = `
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

		CREATE INDEX idx_referrer    ON page_events(referrer);
		CREATE INDEX idx_title 	     ON page_events(title);
		CREATE INDEX idx_remote_addr ON page_events(remote_addr);
	`
)

//reform:page_events
type PageEvent struct {
	ID              int32     `reform:"id,pk"`
	SchemaVersion   int       `reform:"schema_version"  json:"schema_version"`
	ScriptVersion   string    `reform:"script_version"  json:"script_version"`
	Time            time.Time `reform:"datetime"        json:"datetime"`
	Host            string    `reform:"server"          json:"server"`
	RemoteAddr      string    `reform:"remote_addr"     json:"remote_addr"`
	UserAgent       string    `reform:"user_agent"      json:"user_agent"`
	RequestReferrer string    `reform:"request_referrer" json:"request_referrer"`
	Title           string    `reform:"title"           json:"title"`
	PageReferrer    string    `reform:"referrer"         json:"referrer"`
	URL             string    `reform:"url"             json:"url"`
	EventType       string    `reform:"event_type"      json:"event_type"`
	SessionToken    string    `reform:"session_token"   json:"session_token"`
	UserToken       string    `reform:"user_token"      json:"user_token"`
}

func PageEventFromRequest(req *http.Request) *PageEvent {
	params := req.URL.Query()
	return &PageEvent{
		SchemaVersion:   PageEventSchemaVersion,
		Host:            req.Host,
		RemoteAddr:      req.RemoteAddr,
		UserAgent:       req.UserAgent(),
		RequestReferrer: req.Referer(),
		Time:            time.Now(),

		Title:         params.Get("title"),
		PageReferrer:  params.Get("referrer"),
		URL:           params.Get("url"),
		ScriptVersion: params.Get("version"),
		EventType:     params.Get("event_type"),
		SessionToken:  params.Get("session_token"),
		UserToken:     params.Get("user_token"),
	}
}
