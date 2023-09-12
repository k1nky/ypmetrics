package storage

import (
	"testing"
	"time"

	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/stretchr/testify/suite"
)

type newStorageSuite struct {
	suite.Suite
	logger storageLogger
}

func (suite *newStorageSuite) SetupTest() {
	suite.logger = &logger.Blackhole{}
}

func (suite *newStorageSuite) TestDefaultStorage() {
	db := NewStorage(Config{}, suite.logger)
	if _, ok := db.(*MemStorage); !ok {
		suite.Failf("", "expected MemStorage, got: %T", db)
	}
}

func (suite *newStorageSuite) TestSyncFileStorage() {
	db := NewStorage(Config{
		StoragePath: "/tmp/metrics.json",
	}, suite.logger)
	if _, ok := db.(*SyncFileStorage); !ok {
		suite.Failf("", "expected SyncFileStorage, got: %T", db)
	}
}

func (suite *newStorageSuite) TestAsyncFileStorage() {
	db := NewStorage(Config{
		StoreInterval: 10 * time.Second,
		StoragePath:   "/tmp/metrics.json",
	}, suite.logger)
	if _, ok := db.(*AsyncFileStorage); !ok {
		suite.Failf("", "expected AsyncFileStorage, got: %T", db)
	}
}

func (suite *newStorageSuite) TestDBStorage() {
	db := NewStorage(Config{
		StoragePath: "/tmp/metrics.json",
		DSN:         "postgres://",
	}, suite.logger)
	if _, ok := db.(*DBStorage); !ok {
		suite.Failf("", "expected DBStorage, got: %T", db)
	}
}

func TestOpenStorage(t *testing.T) {
	suite.Run(t, new(newStorageSuite))
}
