package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGopsInit(t *testing.T) {
	c := Gops{}
	err := c.Init()
	assert.NoError(t, err)
}

func TestGopsCollect(t *testing.T) {
	c := Gops{}
	c.Init()
	m, err := c.Collect(context.TODO())
	assert.NoError(t, err)
	assert.NotZero(t, len(m.Gauges))
	assert.Zero(t, len(m.Counters))
}

func BenchmarkGopsCollect(b *testing.B) {
	b.StopTimer()
	c := Gops{}
	c.Init()
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		c.Collect(ctx)
	}
}
