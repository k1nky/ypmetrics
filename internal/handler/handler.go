// Пакет handler реализует обработчик HTTP запросов к серверу сбора метрик
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/metricset/server"
)

// typeMetric тип метрики
type typeMetric string

const (
	TypeGauge   = typeMetric("gauge")
	TypeCounter = typeMetric("counter")
)

// IsValid возвращает true, если тип метрики имеет допустимое значение.
func (t typeMetric) IsValid() bool {
	switch t {
	case TypeGauge, TypeCounter:
		return true
	default:
		return false
	}
}

// isValidMetricName возвращает true, если имя метрики имеет допустимое значение.
func isValidMetricName(s string) bool {
	return len(s) > 0
}

// Обработчик вывода всех метрик на сервере
func AllMetricsHandler(ms server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := ms.GetMetrics()
		result := strings.Builder{}
		for _, m := range metrics.Counters {
			result.WriteString(fmt.Sprintf("%s = %s\n", m.Name, m))
		}
		for _, m := range metrics.Gauges {
			result.WriteString(fmt.Sprintf("%s = %s\n", m.Name, m))
		}
		c.String(http.StatusOK, result.String())
	}
}

// Обработчик вывода текущего значения запрашиваемой метрики
func ValueHandler(ms server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := typeMetric(c.Param("type"))
		if !t.IsValid() || !isValidMetricName(c.Param("name")) {
			c.String(http.StatusNotFound, "")
			return
		}
		strValue := ""
		switch t {
		case TypeCounter:
			m := ms.GetCounter(c.Param("name"))
			if m == nil {
				c.String(http.StatusNotFound, "")
				return
			}
			strValue = m.String()
		case TypeGauge:
			m := ms.GetGauge(c.Param("name"))
			if m == nil {
				c.String(http.StatusNotFound, "")
				return
			}
			strValue = m.String()
		}
		c.String(http.StatusOK, strValue)
	}
}

// Обработчик обновления значения указаной метрики
func UpdateHandler(ms server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := typeMetric(c.Param("type"))
		if !t.IsValid() || !isValidMetricName(c.Param("name")) {
			c.String(http.StatusBadRequest, "")
			return
		}
		switch t {
		case TypeCounter:
			if v, err := convertToInt64(c.Param("value")); err != nil {
				c.Status(http.StatusBadRequest)
				return
			} else {
				ms.UpdateCounter(c.Param("name"), v)
			}
		case TypeGauge:
			if v, err := convertToFloat64(c.Param("value")); err != nil {
				c.Status(http.StatusBadRequest)
				return
			} else {
				ms.UpdateGauge(c.Param("name"), v)
			}
		}
		c.Status(http.StatusOK)
	}
}

func convertToInt64(s string) (v int64, err error) {
	v, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return
}

func convertToFloat64(s string) (v float64, err error) {
	v, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return
}
