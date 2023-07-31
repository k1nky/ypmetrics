package handler

import (
	"errors"
	"net/http"

	"github.com/k1nky/ypmetrics/internal/metric"
	"github.com/k1nky/ypmetrics/internal/server"

	"github.com/gin-gonic/gin"
)

type handler struct {
	srv *server.Server
}

func New(srv *server.Server) http.Handler {
	h := &handler{
		srv: srv,
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	valueRoutes := router.Group("/value")
	valueRoutes.GET("/:type/:name", h.valueHandler)
	updateRoutes := router.Group("/update")
	updateRoutes.POST("/:type/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	updateRoutes.POST("/:type/:name/:value", h.updateHandler)

	return router
}

func (h *handler) updateHandler(c *gin.Context) {
	m, err := metric.NewWtihValue(metric.Type(c.Param("type")), c.Param("name"), c.Param("value"))
	if errors.Is(err, metric.ErrEmptyName) {
		c.String(http.StatusNotFound, "%s", err)
		return
	}
	if err != nil {
		c.String(http.StatusBadRequest, "valid format: /update/<type>/<name>/<value>\n")
		return
	}
	h.srv.UpdateMetric(m)
	c.Status(http.StatusOK)
}

func (h *handler) valueHandler(c *gin.Context) {
	m, err := h.srv.GetMetric(metric.Type(c.Param("type")), c.Param("name"))
	if err != nil {
		c.String(http.StatusBadRequest, "valid format: /value/<type>/<name>\n")
		return
	}
	if m == nil {
		c.String(http.StatusNotFound, "")
		return
	}
	c.String(http.StatusOK, m.StringValue())
}
