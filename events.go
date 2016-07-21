//go:generate reform

package main

import (
	"encoding/json"
	"fmt"
	// "gopkg.in/reform.v1"
	"net/http"
	"os"
	"time"
)

const (
	PageEventSchemaVersion = "1"
)

//reform:page_events
type PageEvent struct {
	SchemaVersion  string    `reform:"schema_version"  json:"schema_version"`
	ScriptVersion  string    `reform:"script_version"  json:"script_version"`
	Time           time.Time `reform:"datetime"        json:"datetime"`
	Host           string    `reform:"server"          json:"server"`
	RemoteAddr     string    `reform:"remote_addr"     json:"remote_addr"`
	UserAgent      string    `reform:"user_agent"      json:"user_agent"`
	RequestReferer string    `reform:"request_referer" json:"request_referer"`
	Title          string    `reform:"title"           json:"title"`
	PageReferrer   string    `reform:"referer"         json:"referer"`
	URL            string    `reform:"url"             json:"url"`
	EventType      string    `reform:"event_type"      json:"event_type"`
	SessionToken   string    `reform:"session_token"   json:"session_token"`
	UserToken      string    `reform:"user_token"      json:"user_token"`
}

func PageEventFromRequest(req *http.Request) *PageEvent {
	params := req.URL.Query()
	return &PageEvent{
		SchemaVersion:  PageEventSchemaVersion,
		Host:           req.Host,
		RemoteAddr:     req.RemoteAddr,
		UserAgent:      req.UserAgent(),
		RequestReferer: req.Referer(),
		Time:           time.Now(),

		Title:         params.Get("title"),
		PageReferrer:  params.Get("referer"),
		URL:           params.Get("url"),
		ScriptVersion: params.Get("version"),
		EventType:     params.Get("event_type"),
		SessionToken:  params.Get("session_token"),
		UserToken:     params.Get("user_token"),
	}
}

func LogPageEvent(req *http.Request, writeQueue chan []byte) {
	event := PageEventFromRequest(req)
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create JSON data from event.")
	} else {
		writeQueue <- data
	}
}
