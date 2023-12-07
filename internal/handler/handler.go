// Пакет handler реализует обработчик HTTP запросов к серверу сбора метрик
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	"github.com/k1nky/ypmetrics/internal/protocol"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
)

// Обработчик http-запросов к серверу метрик
type Handler struct {
	keeper keeper.Keeper
}

// New возвращает новый обработчик http-запросов к серверу метрик
func New(keeper keeper.Keeper) Handler {
	return Handler{
		keeper: keeper,
	}
}

// AllMetrics обработчик вывода всех метрик на сервере.
func (h Handler) AllMetrics() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics := metric.Metrics{}
		if err := h.keeper.Snapshot(ctx.Request.Context(), &metrics); err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
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

// Value Обработчик вывода текущего значения запрашиваемой метрики.
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
			m := h.keeper.GetCounter(ctx.Request.Context(), ctx.Param("name"))
			if m == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			strValue = m.String()
		case TypeGauge:
			m := h.keeper.GetGauge(ctx.Request.Context(), ctx.Param("name"))
			if m == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			strValue = m.String()
		}
		ctx.String(http.StatusOK, strValue)
	}
}

// ValueJSON Обработчик вывода текущего значения запрашиваемой метрики в формате JSON.
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
			mm := h.keeper.GetCounter(ctx.Request.Context(), m.ID)
			if mm == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			m.Delta = &mm.Value
		case TypeGauge:
			mm := h.keeper.GetGauge(ctx.Request.Context(), m.ID)
			if mm == nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			m.Value = &mm.Value
		}
		ctx.JSON(http.StatusOK, m)
	}
}

// Update Обработчик обновления значения указаной метрики.
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
				if err := h.keeper.UpdateCounter(ctx.Request.Context(), ctx.Param("name"), v); err != nil {
					ctx.Status(http.StatusInternalServerError)
					return
				}
			}
		case TypeGauge:
			if v, err := convertToFloat64(ctx.Param("value")); err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			} else {
				if err := h.keeper.UpdateGauge(ctx.Request.Context(), ctx.Param("name"), v); err != nil {
					ctx.Status(http.StatusInternalServerError)
					return
				}
			}
		}
		ctx.Status(http.StatusOK)
	}
}

// UpdateJSON Обработчик обновления значения указаной метрики из JSON.
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
			if err := h.keeper.UpdateCounter(ctx.Request.Context(), m.ID, *m.Delta); err != nil {
				ctx.Status(http.StatusInternalServerError)
				return
			}
			c := h.keeper.GetCounter(ctx.Request.Context(), m.ID)
			m.Delta = &c.Value
		case TypeGauge:
			if m.Value == nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
			if err := h.keeper.UpdateGauge(ctx.Request.Context(), m.ID, *m.Value); err != nil {
				ctx.Status(http.StatusInternalServerError)
				return
			}
			g := h.keeper.GetGauge(ctx.Request.Context(), m.ID)
			m.Value = &g.Value
		}
		ctx.JSON(http.StatusOK, m)
	}
}

// Ping Обработчки проверки подключения к БД.
func (h Handler) Ping() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), time.Second)
		defer cancel()
		if err := h.keeper.Ping(c); err != nil {
			ctx.Status(http.StatusInternalServerError)
		} else {
			ctx.Status(http.StatusOK)
		}
	}
}

// UpdatesJSON Обработчик обновления метрик из JSON.
func (h Handler) UpdatesJSON() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		recievedMetrics := make([]protocol.Metrics, 0, 10)
		if err := json.NewDecoder(ctx.Request.Body).Decode(&recievedMetrics); err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		metrics := metric.NewMetrics()
		for _, m := range recievedMetrics {
			switch metricType(m.MType) {
			case TypeCounter:
				metrics.Counters = append(metrics.Counters, metric.NewCounter(m.ID, *m.Delta))
			case TypeGauge:
				metrics.Gauges = append(metrics.Gauges, metric.NewGauge(m.ID, *m.Value))
			}
		}
		if err := h.keeper.UpdateMetrics(ctx.Request.Context(), *metrics); err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Status(http.StatusOK)
	}
}
