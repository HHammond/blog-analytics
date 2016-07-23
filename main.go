package main

import (
	"database/sql"
	"encoding/base64"
	"github.com/caarlos0/env"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	pixelRaw   = "R0lGODlhAQABAIAAANvf7wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=="
	pixelRoute = "/a.gif"
)

type Config struct {
	Port                 string `env:"PORT"             envDefault:"8080"`
	DBFile               string `env:"DB_FILE"          envDefault:"events.db"`
	MaxConnections       int    `env:"MAX_CONNECTIONS"  envDefault:"500000"`
	WriteQueueSize       int    `env:"WRITE_QUEUE_SIZE" envDefault:"100000"`
	WriteFrequencyMillis int    `env:"WRITE_FREQUENCY"  envDefault:"15"`
}

func main() {
	var config Config
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	db := initDB(&config)
	defer db.Close()

	eventQueue := make(chan *PageEvent, config.MaxConnections)

	go writeEvents(&config, db, eventQueue)
	http.HandleFunc(pixelRoute, makeHandlePixel(eventQueue))

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func makeHandlePixel(eventQueue chan *PageEvent) http.HandlerFunc {
	pixel, err := base64.StdEncoding.DecodeString(pixelRaw)
	if err != nil {
		log.Fatalf("failed to encode pixel: %v", err)
	}

	return func(out http.ResponseWriter, req *http.Request) {
		eventQueue <- PageEventFromRequest(req)
		out.Header().Set("Content-Type", "image/gif")
		out.WriteHeader(http.StatusOK)
		out.Write(pixel)
	}
}

func writeEvents(config *Config, db *sql.DB, eventQueue chan *PageEvent) {
	writeFrequency := time.Duration(time.Duration(config.WriteFrequencyMillis) * time.Millisecond)
	writeTimer := time.NewTimer(writeFrequency)
	writeQueue := make([]*PageEvent, 0, config.WriteQueueSize)

	for {
		select {
		case ev := <-eventQueue:
			writeQueue = append(writeQueue, ev)

		case <-writeTimer.C:
			if len(writeQueue) > 0 {
				writeTimer.Stop()

				db.Exec("BEGIN TRANSACTION;")
				for _, ev := range writeQueue {
					ev.InsertIntoDB(db)
				}
				db.Exec("END TRANSACTION;")

				log.Println("wrote", len(writeQueue), "records.")
				writeQueue = writeQueue[:0]
			}
			writeTimer.Reset(writeFrequency)
		}
	}
}

func initDB(config *Config) *sql.DB {
	db, err := sql.Open("sqlite3", config.DBFile)

	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("database is null")
	}

	_, err = db.Exec(SQLPageEventCreateTable)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
