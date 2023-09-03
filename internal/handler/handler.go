// Пакет handler реализует обработчик HTTP запросов к серверу сбора метрик
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/protocol"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
)

// Обработчик запросов к REST API набора метрик
type Handler struct {
	keeper keeper.Keeper
}

func New(keeper keeper.Keeper) Handler {
	return Handler{
		keeper: keeper,
	}
}

// Обработчик вывода всех метрик на сервере
func (h Handler) AllMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics := metric.Metrics{}
		h.keeper.Snapshot(&metrics)
		result := strings.Builder{}
		for _, m := range metrics.Counters {
			result.WriteString(fmt.Sprintf("%s = %s\n", m.Name, m))
		}
		for _, m := range metrics.Gauges {
			result.WriteString(fmt.Sprintf("%s = %s\n", m.Name, m))
		}
		ctx.Writer.Header().Add("content-type", "text/html")
		ctx.String(http.StatusOK, result.String())
	}
}

// Обработчик вывода текущего значения запрашиваемой метрики
func (h Handler) Value() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		t := metricType(ctx.Param("type"))
		if !isValidMetricParams(t, ctx.Param("name")) {
			ctx.Status(http.StatusBadRequest)
			return
		}
		strValue := ""
		switch t {
		case TypeCounter:
			m := h.keeper.GetCounter(ctx.Param("name"))
			if m == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			strValue = m.String()
		case TypeGauge:
			m := h.keeper.GetGauge(ctx.Param("name"))
			if m == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			strValue = m.String()
		}
		ctx.String(http.StatusOK, strValue)
	}
}

// Обработчик вывода текущего значения запрашиваемой метрики в формате JSON
func (h Handler) ValueJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var m protocol.Metrics
		if err := json.NewDecoder(ctx.Request.Body).Decode(&m); err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		t := metricType(m.MType)
		if !isValidMetricParams(t, m.ID) {
			ctx.Status(http.StatusBadRequest)
			return
		}
		switch t {
		case TypeCounter:
			mm := h.keeper.GetCounter(m.ID)
			if mm == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			m.Delta = &mm.Value
		case TypeGauge:
			mm := h.keeper.GetGauge(m.ID)
			if mm == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			m.Value = &mm.Value
		}
		ctx.JSON(http.StatusOK, m)
	}
}

// Обработчик обновления значения указаной метрики
func (h Handler) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		t := metricType(ctx.Param("type"))
		if !isValidMetricParams(t, ctx.Param("name")) {
			ctx.Status(http.StatusBadRequest)
			return
		}
		switch t {
		case TypeCounter:
			if v, err := convertToInt64(ctx.Param("value")); err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			} else {
				h.keeper.UpdateCounter(ctx.Param("name"), v)
			}
		case TypeGauge:
			if v, err := convertToFloat64(ctx.Param("value")); err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			} else {
				h.keeper.UpdateGauge(ctx.Param("name"), v)
			}
		}
		ctx.Status(http.StatusOK)
	}
}

// Обработчик обновления значения указаной метрики из JSON
func (h Handler) UpdateJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var m protocol.Metrics
		if err := json.NewDecoder(ctx.Request.Body).Decode(&m); err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		t := metricType(m.MType)
		if !isValidMetricParams(t, m.ID) {
			ctx.Status(http.StatusBadRequest)
			return
		}
		switch t {
		case TypeCounter:
			if m.Delta == nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
			h.keeper.UpdateCounter(m.ID, *m.Delta)
			c := h.keeper.GetCounter(m.ID)
			m.Delta = &c.Value
		case TypeGauge:
			if m.Value == nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
			h.keeper.UpdateGauge(m.ID, *m.Value)
			g := h.keeper.GetGauge(m.ID)
			m.Value = &g.Value
		}
		ctx.JSON(http.StatusOK, m)
	}
}

// Обработчки проверки подключения к БД
func (h Handler) Ping() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h.keeper.Ping(); err == nil {
			ctx.Status(http.StatusOK)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
	}
}
