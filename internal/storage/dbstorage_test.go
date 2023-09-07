package storage

import (
	"context"
	"os"
	"testing"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type dbStorageTestSuite struct {
	suite.Suite
	db *DBStorage
}

func (suite *dbStorageTestSuite) SetupTest() {
	suite.db = NewDBStorage(&logger.Blackhole{})
	if err := suite.db.Open("postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"); err != nil {
		suite.FailNow(err.Error())
		return
	}
	suite.db.Exec(`TRUNCATE counter;`)
	suite.db.Exec(`TRUNCATE gauge;`)
}

func (suite *dbStorageTestSuite) shouldSkip() {
	if len(os.Getenv("TEST_DB_READY")) == 0 {
		suite.T().Skip()
	}
}

func TestDBStorage(t *testing.T) {
	suite.Run(t, new(dbStorageTestSuite))
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateCounter() {
	suite.shouldSkip()
	ctx := context.TODO()
	suite.db.UpdateCounter(ctx, "c0", 1)
	suite.db.UpdateCounter(ctx, "c0", 10)
	m := suite.db.GetCounter(ctx, "c0")
	assert.Equal(suite.T(), metric.NewCounter("c0", 11), m)
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateGauge() {
	suite.shouldSkip()
	ctx := context.TODO()
	suite.db.UpdateGauge(ctx, "g0", 1)
	suite.db.UpdateGauge(ctx, "g0", 752304.097156)
	m := suite.db.GetGauge(ctx, "g0")
	assert.Equal(suite.T(), metric.NewGauge("g0", 752304.097156), m)
}

func (suite *dbStorageTestSuite) TestDBStorageUpdateMetrics() {
	suite.shouldSkip()
	ctx := context.TODO()
	m := metric.Metrics{
		Counters: []*metric.Counter{metric.NewCounter("c0", 1), metric.NewCounter("c1", 10)},
		Gauges:   []*metric.Gauge{metric.NewGauge("g0", 2.2)},
	}
	suite.db.UpdateMetrics(ctx, m)
	snapshot := metric.NewMetrics()
	suite.db.Snapshot(ctx, snapshot)
	assert.ElementsMatch(suite.T(), m.Counters, snapshot.Counters)
	assert.ElementsMatch(suite.T(), m.Gauges, snapshot.Gauges)
}
