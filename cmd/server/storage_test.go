package main

import (
	"testing"

	"github.com/k1nky/ypmetrics/internal/config"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/storage"
	"github.com/stretchr/testify/suite"
)

type openStorageSuite struct {
	suite.Suite
	logger *logger.Logger
}

func (suite *openStorageSuite) SetupTest() {
	suite.logger = logger.New()
}

func (suite *openStorageSuite) TestDefaultStorage() {
	cfg := config.KeeperConfig{}
	db, _ := openStorage(cfg, suite.logger)
	if _, ok := db.(*storage.MemStorage); !ok {
		suite.Failf("", "expected MemStorage, got: %T", db)
	}
}

func (suite *openStorageSuite) TestSyncFileStorage() {
	cfg := config.KeeperConfig{
		StoreIntervalInSec: 0,
		FileStoragePath:    "/tmp/metrics.json",
	}
	db, _ := openStorage(cfg, suite.logger)
	if _, ok := db.(*storage.SyncFileStorage); !ok {
		suite.Failf("", "expected SyncFileStorage, got: %T", db)
	}
}

func (suite *openStorageSuite) TestAsyncFileStorage() {
	cfg := config.KeeperConfig{
		StoreIntervalInSec: 10,
		FileStoragePath:    "/tmp/metrics.json",
	}
	db, _ := openStorage(cfg, suite.logger)
	if _, ok := db.(*storage.AsyncFileStorage); !ok {
		suite.Failf("", "expected AsyncFileStorage, got: %T", db)
	}
}

func (suite *openStorageSuite) TestDBStorage() {
	cfg := config.KeeperConfig{
		FileStoragePath: "/tmp/metrics.json",
		DatabaseDSN:     "postgres://",
	}
	db, _ := openStorage(cfg, suite.logger)
	if _, ok := db.(*storage.DBStorage); !ok {
		suite.Failf("", "expected DBStorage, got: %T", db)
	}
}

func TestOpenStorage(t *testing.T) {
	suite.Run(t, new(openStorageSuite))
}
