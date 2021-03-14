package recovery

import (
  "github.com/gin-gonic/gin"
  "github.com/tlowerison/go-service/middleware"
)

type Middleware struct {}

func New() *Middleware {
  return &Middleware{}
}

func (*Middleware) Register() {}

func (*Middleware) Handler() gin.HandlerFunc {
  return func(c *gin.Context) {
    middleware.SetStart(c)
    defer func() {
      if err := recover(); err != nil {
        c.AbortWithError(500, err.(error))
      }
    }()
    c.Next()
  }
}
