package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/caarlos0/env"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"time"
)

const (
	pixelRaw   = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
	pixelRoute = "/a.gif"
)

type Config struct {
	Port                 string `env:"PORT"             envDefault:"8080"`
	DBFile               string `env:"DB_FILE"          envDefault:"out.db"`
	MaxConnections       int    `env:"MAX_CONNECTIONS"  envDefault:"100000"`
	WriteQueueSize       int    `env:"WRITE_QUEUE_SIZE" envDefault:"100000"`
	WriteFrequencyMillis int    `env:"WRITE_FREQUENCY"  envDefault:"8"`
}

var cfg = Config{}

func main() {
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	db := initDB()
	defer db.Close()

	eventQueue := make(chan *PageEvent, cfg.MaxConnections)
	go runEventWriter(db, eventQueue)

	http.HandleFunc(pixelRoute, makeHandlePixel(eventQueue))
	if http.ListenAndServe(":"+cfg.Port, nil) != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server")
	}
}

func makeHandlePixel(eventQueue chan *PageEvent) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		panic(err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		eventQueue <- PageEventFromRequest(req)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func runEventWriter(db *sql.DB, eventQueue chan *PageEvent) {
	queue := make([]*PageEvent, 0, cfg.WriteQueueSize)
	executeWrite := make(chan bool)

	writeFrequency := time.Duration(cfg.WriteFrequencyMillis)
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * writeFrequency):
				executeWrite <- true
			}
		}
	}()

	for {
		select {
		case ev := <-eventQueue:
			queue = append(queue, ev)
		case <-executeWrite:
			if len(queue) > 0 {
				db.Exec("BEGIN TRANSACTION;")
				for _, ev := range queue {
					ev.InsertIntoDB(db)
				}
				db.Exec("END TRANSACTION;")
				fmt.Println("Wrote", len(queue), "records.")
				queue = queue[:0]
			}
		}
	}
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", cfg.DBFile)

	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("DB is null")
	}

	_, err = db.Exec(SQLPageEventCreateTable)
	if err != nil {
		panic(err)
	}

	return db
}
