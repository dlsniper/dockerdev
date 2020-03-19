package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	// Set the flags for the logging package to give us the filename in the logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbPool := getDBConnection(context.Background())
	defer dbPool.Close()

	log.Println("starting server...")
	http.HandleFunc("/", homeHandler(dbPool))
	http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		visitorID := 0
		err := dbPool.QueryRow(r.Context(), "INSERT INTO visitors(user_agent, datetime) VALUES ($1, now()) RETURNING id", r.UserAgent()).Scan(&visitorID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to update the database"))
			log.Printf("[error] dbPool error: %v\n", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `Hello, visitor %d!`, visitorID)
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func homeHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		visitorID := 0
		//language=sql
		err := db.QueryRow(r.Context(), "INSERT INTO visitors(user_agent, datetime) VALUES ($1, now()) RETURNING id", r.UserAgent()).Scan(&visitorID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to update the database"))
			log.Printf("[or] db error: %v\n", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `Hello, visitor %d!`, visitorID)
	}
}

func getDBConnection(ctx context.Context) *pgxpool.Pool {
	// Retrieve the database host address
	host := os.Getenv("DD_DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	const connectionString = "postgres://goland:goland@%s:5432/goland?sslmode=disable"

	// Try connecting to the database a few times before giving up
	// Retry to connect for a while
	var dbPool *pgxpool.Pool
	var err error
	for i := 1; i < 8; i++ {
		log.Printf("trying to connect to the db server (attempt %d)...\n", i)
		dbPool, err = pgxpool.Connect(ctx, fmt.Sprintf(connectionString, host))
		if err == nil {
			break
		}
		log.Printf("got error: %v\n", err)

		// Sleep a bit before trying again
		time.Sleep(time.Duration(i*i) * time.Second)
	}

	// Stop execution if the database was not initialized
	if dbPool == nil {
		log.Fatalln("could not connect to the database")
	}

	// Get a connection from the pool and check if the database connection is active and working
	db, err := dbPool.Acquire(ctx)
	if err != nil {
		log.Fatalf("failed to get connection on startup: %v\n", err)
	}
	if err := db.Conn().Ping(ctx); err != nil {
		log.Fatalln(err)
	}

	// Add the connection back to the pool
	db.Release()

	return dbPool
}
