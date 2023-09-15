package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/retrier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type dbStorageTestSuite struct {
	suite.Suite
	db *DBStorage
}

func openTestDB() (*DBStorage, error) {
	db := NewDBStorage(&logger.Blackhole{}, retrier.New())
	cfg := Config{
		DSN: "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable",
	}
	if err := db.Open(cfg); err != nil {
		return nil, err
	}
	db.Exec(`TRUNCATE counter;`)
	db.Exec(`TRUNCATE gauge;`)
	return db, nil
}

func (suite *dbStorageTestSuite) SetupTest() {
	if shouldSkipDBTest(suite.T()) {
		return
	}
	var err error
	if suite.db, err = openTestDB(); err != nil {
		suite.FailNow(err.Error())
		return
	}
}

func shouldSkipDBTest(t *testing.T) bool {
	if len(os.Getenv("TEST_DB_READY")) == 0 {
		t.Skip()
		return true
	}
	return false
}

func TestDBStorage(t *testing.T) {
	suite.Run(t, new(dbStorageTestSuite))
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateCounter() {
	ctx := context.TODO()
	suite.db.UpdateCounter(ctx, "c0", 1)
	suite.db.UpdateCounter(ctx, "c0", 10)
	m := suite.db.GetCounter(ctx, "c0")
	assert.Equal(suite.T(), metric.NewCounter("c0", 11), m)
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateGauge() {
	ctx := context.TODO()
	suite.db.UpdateGauge(ctx, "g0", 1)
	suite.db.UpdateGauge(ctx, "g0", 752304.097156)
	m := suite.db.GetGauge(ctx, "g0")
	assert.Equal(suite.T(), metric.NewGauge("g0", 752304.097156), m)
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateMetrics() {
	ctx := context.TODO()
	m := metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("c0", 1), metric.NewCounter("c1", 10)},
		Gauges:   []*metric.Gauge{metric.NewGauge("g0", 2.2), metric.NewGauge("g1", 12.15)},
	}
	suite.db.UpdateMetrics(ctx, m)
	snapshot := metric.NewMetrics()
	suite.db.Snapshot(ctx, snapshot)
	assert.ElementsMatch(suite.T(), m.Counters, snapshot.Counters)
	assert.ElementsMatch(suite.T(), m.Gauges, snapshot.Gauges)
}

func generateMetrics(metrics *metric.Metrics, count int) {
	if metrics == nil {
		return
	}
	for i := 0; i < count; i++ {
		metrics.Counters = append(metrics.Counters, metric.NewCounter(fmt.Sprintf("counter_%d", i), int64(i)))
	}
}
func BenchmarkBulkUpdateWithUnnet(b *testing.B) {

	ctx := context.TODO()
	fail := func(err error) {
		b.Error(err)
		b.FailNow()
	}
	db, err := openTestDB()
	if err != nil {
		fail(err)
	}
	defer db.Close()

	batches := []int{10, 100, 1000, 10000}

	for _, n := range batches {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			db.Exec(`TRUNCATE counter;`)
			for i := 0; i < b.N; i++ {
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					fail(err)
				}

				metrics := &metric.Metrics{}
				generateMetrics(metrics, n)

				b.ResetTimer()
				stmt, err := tx.PrepareContext(ctx, `
				INSERT INTO counter as c (name, value)
				VALUES (UNNEST($1::varchar[]), UNNEST($2::bigint[]))
				ON CONFLICT ON CONSTRAINT counter_name_key
				DO UPDATE SET value = c.value + EXCLUDED.value
				`)
				if err != nil {
					fail(err)
				}
				defer stmt.Close()
				names := make([]string, 0, len(metrics.Counters))
				values := make([]int64, 0, len(metrics.Counters))
				for _, m := range metrics.Counters {
					names = append(names, m.Name)
					values = append(values, m.Value)
				}
				if _, err := stmt.ExecContext(ctx, names, values); err != nil {
					fail(err)
				}
				if err := tx.Commit(); err != nil {
					fail(err)
				}
			}
		})
	}
}

func BenchmarkBulkUpdateWithLoop(b *testing.B) {

	ctx := context.TODO()
	fail := func(err error) {
		b.Error(err)
		b.FailNow()
	}
	db, err := openTestDB()
	if err != nil {
		fail(err)
	}
	defer db.Close()

	batches := []int{10, 100, 1000, 10000}

	for _, n := range batches {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			db.Exec(`TRUNCATE counter;`)
			for i := 0; i < b.N; i++ {
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					fail(err)
				}

				metrics := &metric.Metrics{}
				generateMetrics(metrics, n)

				b.ResetTimer()
				stmt, err := tx.PrepareContext(ctx, `
				INSERT INTO counter as c (name, value)
				VALUES ($1, $2)
				ON CONFLICT ON CONSTRAINT counter_name_key
				DO UPDATE SET value = c.value + EXCLUDED.value
				`)
				if err != nil {
					fail(err)
				}
				defer stmt.Close()
				for _, m := range metrics.Counters {
					if _, err := stmt.ExecContext(ctx, m.Name, m.Value); err != nil {
						fail(err)
					}
				}
				if err := tx.Commit(); err != nil {
					fail(err)
				}
			}
		})
	}
}

func BenchmarkBulkUpdateWithMultipleValues(b *testing.B) {

	ctx := context.TODO()
	fail := func(err error) {
		b.Error(err)
		b.FailNow()
	}
	db, err := openTestDB()
	if err != nil {
		fail(err)
	}
	defer db.Close()

	batches := []int{10, 100, 1000, 10000}

	for _, n := range batches {
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			db.Exec(`TRUNCATE counter;`)
			for i := 0; i < b.N; i++ {

				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					fail(err)
				}
				metrics := &metric.Metrics{}
				generateMetrics(metrics, n)

				b.ResetTimer()
				stmt := `
				INSERT INTO counter as c (name, value)
				VALUES %s
				ON CONFLICT ON CONSTRAINT counter_name_key
				DO UPDATE SET value = c.value + EXCLUDED.value
				`
				args := make([]interface{}, 0, len(metrics.Counters))
				holders := make([]string, 0, len(metrics.Counters))
				for i, m := range metrics.Counters {
					args = append(args, m.Name, m.Value)
					holders = append(holders, fmt.Sprintf("($%d, $%d)", 2*i+1, 2*i+2))
				}
				s := fmt.Sprintf(stmt, strings.Join(holders, ","))

				if _, err := tx.ExecContext(ctx, s, args...); err != nil {
					fail(err)
				}
				if err := tx.Commit(); err != nil {
					fail(err)
				}
			}
		})
	}
}
