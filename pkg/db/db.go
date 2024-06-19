package db

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	userName = "postgres"
	password = "password"

	// we refer to the name specified in compose.yaml for that service.
	host = "db"

	port   = 5432
	dbName = "homework"

	defaultMaxConnAttempts          = 10
	defaultConnectionTickerDuration = 5 * time.Second
)

//go:embed query.txt
var sqlQuert string

type Service struct {
	dbPool *pgxpool.Pool
}

func DefaultConnString() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", userName, password, host, port, dbName)
}

// tryToConnect will try to connect to the "connString" every "d" units of time, a
// "maxConnAttempts" number of times at maximum and create a connection pool.
// It returns a connection pool to the DB and an error, if any.
func tryToCreatePoolAndConnect(connString string, d time.Duration, maxConnAttempts int) (*pgxpool.Pool, error) {
	// From https://pkg.go.dev/github.com/jackc/pgx/v5@v5.6.0/pgxpool
	// "A pool returns without waiting for any connections to be established.
	// Acquire a connection immediately after creating the pool to check if a connection can successfully be established.
	ticker := time.NewTicker(d)
	counter := 0
	var err error

	for range ticker.C {
		if counter >= maxConnAttempts {
			break
		}

		dbPool, err := pgxpool.New(context.Background(), connString)
		// if we can create the connection pool
		if err == nil {
			connection, err := dbPool.Acquire(context.Background())
			if err == nil {
				// we can connect!
				connection.Release()
				return dbPool, nil
			}
		}

		counter++
	}

	return nil, fmt.Errorf("unable to create connection pool, maxConnAttempts tried: %v", err)
}

func NewDBService(connString string) (*Service, error) {
	dbPool, err := tryToCreatePoolAndConnect(connString, defaultConnectionTickerDuration, defaultMaxConnAttempts)
	if err != nil {
		return nil, err
	}

	dbService := &Service{
		dbPool: dbPool,
	}

	return dbService, nil
}

// CountAllRows was used to play around.
func (d *Service) CountAllRows() (int, error) {
	// this query will return all rows
	rawQuery := "select * from cpu_usage"

	// can use Query directly because we don't want measure duration
	// in this example function, otherwise manually handle the connections.
	rows, err := d.dbPool.Query(context.Background(), rawQuery)
	if err != nil {
		return 0, fmt.Errorf("query failed: %v", err)
	}

	// var ts time.Time
	// var host, usage string
	counter := 0
	for rows.Next() {
		counter++
		// err := rows.Scan(&ts, &host, &usage)
		// if err != nil {
		// 	log.Fatalf("unable to scan row: %v", err)
		// }
		// log.Println(ts, host, usage)
	}

	return counter, nil
}

func (d *Service) Process(ctx context.Context, req *domain.Measurement) (*domain.QueryResultWithTime, error) {
	// not using p.dbPool.Query() as that would include the connection "acquire" step,
	// and we want to start measuring the time after we acquire the connection but before we execute the query
	conn, err := d.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	start := time.Now()
	rows, err := conn.Query(ctx, sqlQuert, req.Hostname, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// (maybe we could to this here as the query was executed but we have not scaned the results yet)
	// timeDiff := time.Since(start)

	var queryResults []*domain.QueryResult

	for rows.Next() {
		queryResult := &domain.QueryResult{}
		err := rows.Scan(&queryResult.Timestamp, &queryResult.MaxCPUUsage, &queryResult.MinCPUUsage)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}

		// log.Println("query result is:", *queryResult)

		queryResults = append(queryResults, queryResult)
	}

	timeDiff := time.Since(start)

	return &domain.QueryResultWithTime{
		Queryresults:       queryResults,
		QueryExecutionTime: timeDiff,
	}, nil
}
