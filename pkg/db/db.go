package db

import (
	"context"
	"fmt"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"

	"github.com/jackc/pgx/v5/pgxpool"
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
	dbPool *pgxpool.Pool
}

func DefaultConnString() string {
	return fmt.Sprintf("postgres://%v:%v@%v:%v/%v", userName, password, host, port, dbName)
}

// tryToConnect will try to connect to the "connString" every "d" units of time, a
// "maxConnAttempts" number of times at maximum and create a connection pool.
// It returns a connection pool to the DB and an error, if any.
func tryToCreatePool(connString string, d time.Duration, maxConnAttempts int) (*pgxpool.Pool, error) {
	ticker := time.NewTicker(d)
	i := 0
	var err error

	for range ticker.C {
		if i >= maxConnAttempts {
			break
		}

		dbPool, err := pgxpool.New(context.Background(), connString)
		if err == nil {
			return dbPool, nil
		}

		i++
	}

	return nil, fmt.Errorf("unable to create connection pool, maxConnAttempts tried: %v", err)
}

func NewDBService(connString string) (*DBService, error) {
	dbPool, err := tryToCreatePool(connString, defaultConnectionTickerDuration, defaultMaxConnAttempts)
	if err != nil {
		return nil, err
	}

	dbService := &DBService{
		dbPool: dbPool,
	}

	return dbService, nil
}

func (d *DBService) CountAllRows() (int, error) {
	// rawQuery := "select * from cpu_usage where host = 'host_000000'"
	rawQuery := "select * from cpu_usage"
	rows, err := d.dbPool.Query(context.Background(), rawQuery)
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

func (d *DBService) Process(ctx context.Context, req *domain.Measurement) (*domain.QueryResultWithTime, error) {
	// not using p.dbPool.Query() as this includes the connection acquire,
	// and we want to measure the time after we adquire the connection but before
	// we execute the query!
	conn, err := d.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	rows, err := conn.Query(ctx, "select * from cpu_usage", req.Hostname, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// maybe we could to this here:
	// timeDiff := time.Since(start)

	var queryResults []*domain.QueryResult
	for rows.Next() {
		queryResult := &domain.QueryResult{}
		err := rows.Scan(&queryResult.Timestamp, &queryResult.MaxCpuUsage, &queryResult.MinCpuUsage)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		queryResults = append(queryResults, queryResult)
	}

	timeDiff := time.Since(start)
	return &domain.QueryResultWithTime{
		Queryresults:       queryResults,
		QueryExecutionTime: timeDiff,
	}, nil
}
