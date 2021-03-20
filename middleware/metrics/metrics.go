package metrics

import (
  "github.com/gin-gonic/gin"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"
)

type Cleanup func()
type Handler func(c *gin.Context) Cleanup
type Metric func(registry *prometheus.Registry, factory promauto.Factory) Handler

type Middleware struct {
  Registry *prometheus.Registry
  Factory  promauto.Factory
  Metrics  []Metric
  handlers []Handler
}

func New(metrics []Metric) *Middleware {
  return &Middleware{
    Metrics: metrics,
  }
}

func (m *Middleware) Register() {
  m.Registry = prometheus.NewRegistry()
  m.Factory = promauto.With(m.Registry)
  m.handlers = make([]Handler, len(m.Metrics))
  for i, metric := range m.Metrics {
    m.handlers[i] = metric(m.Registry, m.Factory)
  }
}

func (m *Middleware) Handler() gin.HandlerFunc {
  return func(c *gin.Context) {
    for _, handler := range m.handlers {
      defer handler(c)()
    }
    c.Next()
  }
}
