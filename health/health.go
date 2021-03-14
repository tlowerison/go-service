package health

import (
  "fmt"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/tlowerison/go-service/middleware"
)

type Error struct {
  Err string `json:"error"`
}

type Check func(c *gin.Context) error

func Handler(timeout time.Duration, check Check) gin.HandlerFunc {
  return func(c *gin.Context) {
    start := c.GetTime(middleware.StartKey)
    if (start == time.Time{}) {
      start = time.Now()
    }

    var err error
    select {
    case err = <-waitForCheck(c, check):
    case err = <-waitForTimeout(start, timeout):
    }

    if err != nil {
      c.Error(err)
      c.AbortWithStatusJSON(500, interface{}(Error{Err: err.Error()}))
    } else {
      c.Status(200)
      c.Writer.Write([]byte("ok"))
    }
  }
}

func waitForCheck(c *gin.Context, check Check) <-chan error {
  err := make(chan error)
  go func() {
    defer close(err)
    err <-check(c)
  }()
  return err
}

func waitForTimeout(start time.Time, timeout time.Duration) <-chan error {
  err := make(chan error)
  go func() {
    defer close(err)
    time.Sleep(start.Add(timeout).Sub(time.Now()))
    err <-fmt.Errorf("Request exceeded %v.", timeout)
  }()
  return err
}
