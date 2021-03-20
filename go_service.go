package go_service

import (
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
)

type Registerable interface {
  Register()
}

type Registerables []Registerable

func (registerables Registerables) Register() {
  for _, registerable := range registerables {
    if registerable != nil {
      registerable.Register()
    }
  }
}

type Service interface {
  Serve() <-chan *http.Server
  Registerable
}

type Services []Service

func (services Services) Register() {
  for _, service := range services {
    if service != nil {
      service.Register()
    }
  }
}

type Middleware interface {
  Handler() gin.HandlerFunc
  Registerable
}

type Middlewares []Middleware

func (middlewares Middlewares) Register() {
  for _, middleware := range middlewares {
    if middleware != nil {
      middleware.Register()
    }
  }
}

func (middlewares Middlewares) Apply(router *gin.Engine) {
  for _, middleware := range middlewares {
    if middleware != nil {
      router.Use(middleware.Handler())
    }
  }
}

const StartKey = "start"

func SetStart(c *gin.Context) {
  _, exists := c.Get(StartKey)
  if !exists {
    c.Set(StartKey, interface{}(time.Now()))
  }
}
