package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

// Хранилище метрик в базе данных
type DBStorage struct {
	*sql.DB
	logger storageLogger
}

func NewDBStorage(logger storageLogger) *DBStorage {
	return &DBStorage{
		logger: logger,
	}
}

func (dbs *DBStorage) Open(dataSourceName string) (err error) {
	dbs.DB, err = sql.Open("pgx", dataSourceName)
	if err != nil {
		return err
	}
	return dbs.Init()
}

func (dbs *DBStorage) Init() error {
	tx, err := dbs.Begin()
	if err != nil {
		return err
	}
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS counter (
			id serial PRIMARY KEY,
			name varchar(100),
			value integer,
			UNIQUE (name)
		);
	`)
	tx.Exec(`CREATE TABLE IF NOT EXISTS gauge (
			id serial PRIMARY KEY,
			name varchar(100),
			value double precision,
			UNIQUE (name)
		);
	`)
	return tx.Commit()
}

func (dbs *DBStorage) GetCounter(ctx context.Context, name string) *metric.Counter {
	m := metric.NewCounter(name, 0)
	row := dbs.QueryRowContext(ctx, `SELECT value FROM counter WHERE name=$1`, name)
	if err := row.Err(); err != nil {
		dbs.logger.Error("GetCounter: %v", err)
		return nil
	}
	if err := row.Scan(&m.Value); err != nil {
		if err != sql.ErrNoRows {
			dbs.logger.Error("GetCounter: %v", err)
		}
		return nil
	}
	return m
}

func (dbs *DBStorage) GetGauge(ctx context.Context, name string) *metric.Gauge {
	m := metric.NewGauge(name, 0)
	row := dbs.QueryRowContext(ctx, `SELECT value FROM gauge WHERE name=$1`, name)
	if err := row.Err(); err != nil {
		dbs.logger.Error("GetGauge: %v", err)
		return nil
	}
	if err := row.Scan(&m.Value); err != nil {
		if err != sql.ErrNoRows {
			dbs.logger.Error("GetGauge: %v", err)
		}
		return nil
	}
	return m
}

func (dbs *DBStorage) UpdateCounter(ctx context.Context, name string, value int64) {
	if _, err := dbs.ExecContext(ctx, `
		INSERT INTO counter as c (name, value)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT counter_name_key
		DO UPDATE SET value = c.value + EXCLUDED.value
	`, name, value); err != nil {
		dbs.logger.Error("UpdateCounter: %v", err)
	}
}

func (dbs *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) {
	if _, err := dbs.ExecContext(ctx, `
		INSERT INTO gauge (name, value)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT gauge_name_key
		DO UPDATE SET value = EXCLUDED.value
	`, name, value); err != nil {
		dbs.logger.Error("UpdateGauge: %v", err)
	}
}

func (dbs *DBStorage) Snapshot(ctx context.Context, metrics *metric.Metrics) {

	if metrics == nil {
		return
	}

	counters, err := dbs.QueryContext(ctx, `SELECT name, value FROM counter`)
	if err != nil {
		dbs.logger.Error("Snapshot: %v", err)
		return
	}
	defer counters.Close()
	for counters.Next() {
		m := &metric.Counter{}
		if err := counters.Scan(&m.Name, &m.Value); err != nil {
			dbs.logger.Error("Snapshot: %v", err)
			return
		}
		metrics.Counters = append(metrics.Counters, m)
	}
	if err := counters.Err(); err != nil {
		dbs.logger.Error("Snapshot: %v", err)
		return
	}

	gauges, err := dbs.QueryContext(ctx, `SELECT name, value FROM gauge`)
	if err != nil {
		dbs.logger.Error("Snapshot: %v", err)
		return
	}
	defer gauges.Close()
	for gauges.Next() {
		m := &metric.Gauge{}
		if err := gauges.Scan(&m.Name, &m.Value); err != nil {
			dbs.logger.Error("Snapshot: %v", err)
			return
		}
		metrics.Gauges = append(metrics.Gauges, m)
	}
	if err := gauges.Err(); err != nil {
		dbs.logger.Error("Snapshot: %v", err)
		return
	}
}
