package main

import (
	sqlite "github.com/mattn/go-sqlite3"
	"net/url"
)

func sqlite3Hostname(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return u.Host
}

func registerSQLiteExtensions(conn *sqlite.SQLiteConn) error {
	if err := conn.RegisterFunc("hostname", sqlite3Hostname, true); err != nil {
		return err
	}
	return nil
}
