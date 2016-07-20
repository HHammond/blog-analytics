package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	PageEventSchemaVersion = "1"
)

type PageEvent struct {
	SchemaVersion  string    `json:"schema_version"`
	ScriptVersion  string    `json:"script_version"`
	Time           time.Time `json:"datetime"`
	Host           string    `json:"server"`
	RemoteAddr     string    `json:"remote_addr"`
	UserAgent      string    `json:"user_agent"`
	RequestReferer string    `json:"referer"`
	Title          string    `json:"title"`
	PageReferrer   string    `json:"referer"`
	URL            string    `json:"url"`
	EventType      string    `json:"event_type"`
	SessionToken   string    `json:"session_token"`
	UserToken      string    `json:"user_token"`
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
		fmt.Fprintln(os.Stderr, "Failed to log json data")
	} else {
		writeQueue <- data
	}
}

func WriteEventsToFile(outfile *os.File, messages chan []byte) {
	for msg := range messages {
		n, err := outfile.Write(msg)
		if n != len(msg) || err != nil {
			panic("Failed to write events to file")
		}
	}
}
