package middleware

import (
  "time"

  "github.com/gin-gonic/gin"
)

type Middleware interface {
  Register()
  Handler() gin.HandlerFunc
}

type Collection []Middleware

const StartKey = "start"

func (collection Collection) Register() {
  for _, middleware := range collection {
    if middleware != nil {
      middleware.Register()
    }
  }
}

func (collection Collection) Apply(r *gin.Engine) {
  for _, middleware := range collection {
    if middleware != nil {
      r.Use(middleware.Handler())
    }
  }
}

func SetStart(c *gin.Context) {
  _, exists := c.Get(StartKey)
  if !exists {
    c.Set(StartKey, interface{}(time.Now()))
  }
}
