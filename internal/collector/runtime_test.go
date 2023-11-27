package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeInit(t *testing.T) {
	c := Runtime{}
	err := c.Init()
	assert.NoError(t, err)
}

func TestRuntimeCollect(t *testing.T) {
	c := Runtime{}
	m, err := c.Collect(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, m.Gauges, 27)
	assert.Len(t, m.Counters, 0)
}

func BenchmarkRuntimeCollect(b *testing.B) {
	b.StopTimer()
	c := Runtime{}
	c.Init()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.Collect(ctx)
	}
}
