package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/logger"
	"github.com/k1nky/ypmetrics/internal/retrier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func newTestCounters() map[string]*metric.Counter {
	return map[string]*metric.Counter{
		"c0": metric.NewCounter("c0", 1),
		"c1": metric.NewCounter("c1", 15),
	}
}

func newTestGauges() map[string]*metric.Gauge {
	return map[string]*metric.Gauge{
		"g0": metric.NewGauge("g0", 1.1),
		"g1": metric.NewGauge("g1", 36.6),
	}
}

type fileStorageTestSuite struct {
	suite.Suite
	fs *FileStorage
}

func assertMetricsJSONEq(t assert.TestingT, expected string, actual string) {

	// workaround due to https://github.com/stretchr/testify/issues/1025
	expectedAsMetrics := metric.Metrics{}
	actualAsMetrics := metric.Metrics{}
	if err := json.Unmarshal([]byte(expected), &expectedAsMetrics); err != nil {
		t.Errorf(fmt.Sprintf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error()))
		return
	}

	if err := json.Unmarshal([]byte(actual), &actualAsMetrics); err != nil {
		t.Errorf(fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error()))
		return
	}
	assert.ElementsMatch(t, expectedAsMetrics.Counters, actualAsMetrics.Counters)
	assert.ElementsMatch(t, expectedAsMetrics.Gauges, actualAsMetrics.Gauges)
}

func (suite *fileStorageTestSuite) SetupTest() {
	suite.fs = NewFileStorage(&logger.Blackhole{}, retrier.New())
	suite.fs.counters = newTestCounters()
	suite.fs.gauges = newTestGauges()
}

func (suite *fileStorageTestSuite) TestFlush() {
	buf := bytes.Buffer{}
	if err := suite.fs.Flush(&buf); err != nil {
		suite.T().Errorf("unexpected error = %v", err)
		return
	}
	want := `{"Counters":[{"Name":"c0","Value":1},{"Name":"c1","Value":15}],"Gauges":[{"Name":"g0","Value":1.1},{"Name":"g1","Value":36.6}]}`
	assertMetricsJSONEq(suite.T(), want, buf.String())
}

func (suite *fileStorageTestSuite) TestRestore() {
	buf := bytes.Buffer{}
	buf.WriteString(`{"Counters":[{"Name":"c0","Value":1},{"Name":"c1","Value":15}],"Gauges":[{"Name":"g0","Value":1.1},{"Name":"g1","Value":36.6}]}`)
	suite.fs.counters = make(map[string]*metric.Counter)
	suite.fs.gauges = make(map[string]*metric.Gauge)
	if err := suite.fs.Restore(&buf); err != nil {
		suite.T().Errorf("unexpected error = %v", err)
		return
	}
	wantCounters := newTestCounters()
	wantGauges := newTestGauges()
	suite.Assert().Equal(wantCounters, suite.fs.counters)
	suite.Assert().Equal(wantGauges, suite.fs.gauges)
}

func (suite *fileStorageTestSuite) TestRestoreInvalidJSON() {
	buf := bytes.Buffer{}
	buf.WriteString(`{"Counters":[{"Name":"c0","Value":1},{"Name":"c1","Value":15}],"Gauges":[{"Name":"g0","Value":1.1},{"Name":"g1","Value":36.6}]`)
	if err := suite.fs.Restore(&buf); err == nil {
		suite.T().Errorf("expected error")
	}
}

func (suite *fileStorageTestSuite) TestWriteToFile() {
	filename := "/tmp/123"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		suite.T().Errorf("unexpected error = %v", err)
		return
	}
	defer f.Close()
	if err := suite.fs.WriteToFile(f); err != nil {
		suite.T().Errorf("unexpected error = %v", err)
		return
	}
	ctx := context.TODO()
	data, _ := os.ReadFile(filename)
	want := `{"Counters":[{"Name":"c0","Value":1},{"Name":"c1","Value":15}],"Gauges":[{"Name":"g0","Value":1.1},{"Name":"g1","Value":36.6}]}`
	assertMetricsJSONEq(suite.T(), want, string(data))
	suite.fs.UpdateCounter(ctx, "c2", 20)
	if err := suite.fs.WriteToFile(f); err != nil {
		suite.T().Errorf("unexpected error = %v", err)
		return
	}
	data, _ = os.ReadFile(filename)
	want = `{"Counters":[{"Name":"c0","Value":1},{"Name":"c1","Value":15},{"Name":"c2","Value":20}],"Gauges":[{"Name":"g0","Value":1.1},{"Name":"g1","Value":36.6}]}`
	assertMetricsJSONEq(suite.T(), want, string(data))
}

func TestFileStorage(t *testing.T) {
	suite.Run(t, new(fileStorageTestSuite))
}
