package storage

import (
	"testing"

	"github.com/k1nky/ypmetrics/internal/logger"
)

func TestDBStorage(t *testing.T) {
	t.SkipNow()
	blackholeLogger := &logger.Blackhole{}
	db := NewDBStorage(blackholeLogger)
	if err := db.Open("postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"); err != nil {
		t.Log(err)
		return
	}
	db.UpdateCounter("c0", 1)
	db.UpdateCounter("c0", 10)
	m := db.GetCounter("c0")
	t.Log(m)
}
