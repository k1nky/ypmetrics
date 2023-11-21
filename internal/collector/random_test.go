package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandomInit(t *testing.T) {
	c := Random{}
	err := c.Init()
	assert.NoError(t, err)
}

func TestRandomCollect(t *testing.T) {
	ctx := context.TODO()

	c := Random{}
	c.Init()
	m, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.Len(t, m.Gauges, 1)
	assert.Len(t, m.Counters, 0)
}

func TestRandomCollectSeries(t *testing.T) {
	ctx := context.TODO()

	c := Random{}
	c.Init()
	m1, _ := c.Collect(ctx)
	time.Sleep(time.Second)
	m2, _ := c.Collect(ctx)
	assert.NotEqual(t, m1.Gauges[0].Value, m2.Gauges[0].Value)
}

func BenchmarkRandomCollect(b *testing.B) {
	b.StopTimer()
	c := Random{}
	c.Init()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.Collect(ctx)
	}
}
