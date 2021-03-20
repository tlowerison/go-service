package metrics

import (
  "fmt"
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"
  "github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/rs/zerolog/log"
  flag "github.com/spf13/pflag"
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

type Service struct {
  Middleware *Middleware
  Port       int
}

func New(metrics []Metric) *Service {
  return &Service{Middleware: &Middleware{Metrics: metrics}}
}

func (s *Service) Register() {
  flag.IntVar(&s.Port, "metrics-port", 9090, "Metrics port to listen on.")
}

func (s *Service) Serve() <-chan *http.Server {
  channel := make(chan *http.Server)
  go func() {
    defer close(channel)

    router := gin.New()
    server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: router}

    router.GET("/metrics", s.handler())

  	log.Debug().Msgf("Metrics server starting on port %d.", s.Port)

    channel <- server

  	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
  		log.Fatal().Err(err).Msg("Starting metrics server failed.")
  	}
  }()
  return channel
}

func (s *Service) handler() gin.HandlerFunc {
  return gin.WrapH(promhttp.Handler())
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
