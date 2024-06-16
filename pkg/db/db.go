package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	userName = "postgres"
	password = "password"

	// we refer to the name specified in compose.yaml for that service:
	host = "db"

	port   = 5432
	dbName = "homework"

	defaultMaxConnAttempts          = 10
	defaultConnectionTickerDuration = 5 * time.Second
)

type DBService struct {
	dbConn *pgx.Conn
}

func DefaultConnString() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", userName, password, host, port, dbName)
}

// tryToConnect will try to connect to the "connString" every "d" units of time, a
// "maxConnAttempts" number of times at maximum. It returns a connection to the DB and an error, if any.
func tryToConnect(connString string, d time.Duration, maxConnAttempts int) (*pgx.Conn, error) {
	ticker := time.NewTicker(d)
	i := 0
	var err error
	for range ticker.C {
		if i >= maxConnAttempts {
			break
		}

		dbConn, err := pgx.Connect(context.Background(), connString)
		if err == nil {
			return dbConn, nil
		}
		i++
	}
	return nil, fmt.Errorf("unable to connect to database: %v", err)
}

func NewDBService(connString string) (*DBService, error) {
	dbConn, err := tryToConnect(connString, defaultConnectionTickerDuration, defaultMaxConnAttempts)
	if err != nil {
		log.Fatal(err)
	}

	dbService := &DBService{
		dbConn: dbConn,
	}

	return dbService, nil
}

func (d *DBService) CountAllRows() (int, error) {
	// rawQuery := "select * from cpu_usage where host = 'host_000000'"
	rawQuery := "select * from cpu_usage"
	rows, err := d.dbConn.Query(context.Background(), rawQuery)
	if err != nil {
		return 0, fmt.Errorf("query failed: %v", err)
	}

	// var ts time.Time
	// var host, usage string
	i := 0
	for rows.Next() {
		i++
		// err := rows.Scan(&ts, &host, &usage)
		// if err != nil {
		// 	log.Fatalf("unable to scan row: %v", err)
		// }
		// log.Println(ts, host, usage)
	}
	return i, nil
}
