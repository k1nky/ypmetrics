package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPollCounterInit(t *testing.T) {
	c := PollCounter{}
	err := c.Init()
	assert.NoError(t, err)
}

func TestPollCounterCollect(t *testing.T) {
	ctx := context.TODO()

	c := PollCounter{}
	c.Init()
	m, err := c.Collect(ctx)
	assert.NoError(t, err)
	assert.Len(t, m.Counters, 1)
	assert.Len(t, m.Gauges, 0)
	assert.Equal(t, m.Counters[0].Value, int64(1))
}

func TestPollCounterCollectSeries(t *testing.T) {
	ctx := context.TODO()

	c := PollCounter{}
	c.Init()
	m1, _ := c.Collect(ctx)
	m2, _ := c.Collect(ctx)
	assert.Equal(t, m1.Counters[0], m2.Counters[0])
}

func BenchmarkPollCounterCollect(b *testing.B) {
	b.StopTimer()
	c := PollCounter{}
	c.Init()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.Collect(ctx)
	}
}
