package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
)

const (
	MaxKeepaliveDBConnections = 10
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

// Open открывает подключение к базе данных. Если БД недоступна возвращает ошибку.
// При необходимости выполняет инициализацию базы данных.
func (dbs *DBStorage) Open(dataSourceName string) (err error) {
	dbs.DB, err = sql.Open("pgx", dataSourceName)
	if err != nil {
		return err
	}

	dbs.SetMaxOpenConns(MaxKeepaliveDBConnections)
	dbs.SetMaxIdleConns(MaxKeepaliveDBConnections)

	if err := dbs.Ping(); err != nil {
		return err
	}

	return dbs.Initialize()
}

// Initialize создает схему базы данных
func (dbs *DBStorage) Initialize() error {
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

// GetCounter возвращает метрику Counter по имени name.
// Будет возвращен nil, если метрика не найдена
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

// GetGauge возвращает метрику Gauge по имени name.
// Будет возвращен nil, если метрика не найдена
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

// UpdateCounter обновляет метрику Counter в базе данных
func (dbs *DBStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	if _, err := dbs.ExecContext(ctx, `
		INSERT INTO counter as c (name, value)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT counter_name_key
		DO UPDATE SET value = c.value + EXCLUDED.value
	`, name, value); err != nil {
		dbs.logger.Error("UpdateCounter: %v", err)
		return err
	}
	return nil
}

// UpdateGauge обновляет метрику Gauge в базе данных
func (dbs *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	if _, err := dbs.ExecContext(ctx, `
		INSERT INTO gauge (name, value)
		VALUES ($1, $2)
		ON CONFLICT ON CONSTRAINT gauge_name_key
		DO UPDATE SET value = EXCLUDED.value
	`, name, value); err != nil {
		dbs.logger.Error("UpdateGauge: %v", err)
		return err
	}
	return nil
}

// UpdateMetrics выполняет множественно обнволение метрик. Обновление выполняется в транзакции.
func (dbs *DBStorage) UpdateMetrics(ctx context.Context, metrics metric.Metrics) error {
	// вспомогательная функция, которую вызываем при ошибке
	fail := func(err error) error {
		dbs.logger.Error("UpdateMetrics: %v", err)
		return err
	}

	tx, err := dbs.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	// всегда откатываем изменения, если не выполнился явный Commit
	defer tx.Rollback()
	if len(metrics.Counters) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO counter as c (name, value)
			VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT counter_name_key
			DO UPDATE SET value = c.value + EXCLUDED.value
		`)
		if err != nil {
			return fail(err)
		}
		defer stmt.Close()
		for _, m := range metrics.Counters {
			if _, err := stmt.Exec(m.Name, m.Value); err != nil {
				return fail(err)
			}
		}
	}
	if len(metrics.Gauges) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO gauge as g (name, value)
			VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT gauge_name_key
			DO UPDATE SET value = g.value + EXCLUDED.value
		`)
		if err != nil {
			return fail(err)
		}
		defer stmt.Close()
		for _, m := range metrics.Gauges {
			if _, err := stmt.Exec(m.Name, m.Value); err != nil {
				return fail(err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return nil
}

// Snapshot создает снимок метрик из базы данных
func (dbs *DBStorage) Snapshot(ctx context.Context, metrics *metric.Metrics) error {

	if metrics == nil {
		return nil
	}

	fail := func(err error) error {
		dbs.logger.Error("Snapshot: %v", err)
		return err
	}

	counters, err := dbs.QueryContext(ctx, `SELECT name, value FROM counter`)
	if err != nil {
		return fail(err)
	}
	defer counters.Close()
	for counters.Next() {
		m := &metric.Counter{}
		if err := counters.Scan(&m.Name, &m.Value); err != nil {
			return fail(err)
		}
		metrics.Counters = append(metrics.Counters, m)
	}
	if err := counters.Err(); err != nil {
		return fail(err)
	}

	gauges, err := dbs.QueryContext(ctx, `SELECT name, value FROM gauge`)
	if err != nil {
		return fail(err)
	}
	defer gauges.Close()
	for gauges.Next() {
		m := &metric.Gauge{}
		if err := gauges.Scan(&m.Name, &m.Value); err != nil {
			return fail(err)
		}
		metrics.Gauges = append(metrics.Gauges, m)
	}
	if err := gauges.Err(); err != nil {
		return fail(err)
	}

	return nil
}
