package go_service

import (
  "time"

  "github.com/gin-gonic/gin"
)

type Service interface {
  Register()
  Handler() gin.HandlerFunc
}

type Collection []Service

const StartKey = "start"

func (collection Collection) Register() {
  for _, service := range collection {
    if service != nil {
      service.Register()
    }
  }
}

func (collection Collection) Apply(r *gin.Engine) {
  for _, service := range collection {
    if service != nil {
      r.Use(service.Handler())
    }
  }
}

func SetStart(c *gin.Context) {
  _, exists := c.Get(StartKey)
  if !exists {
    c.Set(StartKey, interface{}(time.Now()))
  }
}
