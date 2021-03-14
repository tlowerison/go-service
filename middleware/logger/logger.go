package logger

import (
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/rs/zerolog"
  "github.com/rs/zerolog/log"
  "github.com/tlowerison/go-service/middleware"
)

type EventAppender func(event *zerolog.Event)

type Middleware struct {
  EventAppender EventAppender
}

func New() *Middleware {
  return &Middleware{
    EventAppender: func(event *zerolog.Event) {},
  }
}

func (*Middleware) Register() {}

func (m *Middleware) Handler() gin.HandlerFunc {
  return func(c *gin.Context) {
    middleware.SetStart(c)
    start := c.GetTime(middleware.StartKey)

    c.Next()

    var event *zerolog.Event
    status := c.Writer.Status()
    if status >= 500 {
      event = log.Error()
    } else {
      event = log.Info()
    }

    params := zerolog.Dict()
    for _, param := range c.Params {
      params.Str(param.Key, param.Value)
    }

    event.
      Str("method", c.Request.Method).
      Str("path", c.Request.RequestURI).
      Int("status", status).
      Str("ip", getIP(c)).
      Float64("duration", float64(time.Now().Sub(start)) / float64(time.Millisecond)).
      Str("referrer", c.Request.Referer()).
      Str("requestId", c.Writer.Header().Get("Request-Id")).
      Dict("params", params)

    if status >= 500 {
      event.Strs("errors", c.Errors.Errors())
    }

    m.EventAppender(event)

    event.Send()
  }
}

func getIP(c *gin.Context) string {
	// first check the X-Forwarded-For header
	requester := c.Request.Header.Get("X-Forwarded-For")
	// if empty, check the Real-IP header
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	// if the requester is still empty, use the hard-coded address from the socket
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	// if requester is a comma delimited list, take the first one
	// (this happens when proxied via elastic load balancer then again through nginx)
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}

	return requester
}
