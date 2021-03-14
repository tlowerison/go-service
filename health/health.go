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

func Handler(timeout time.Duration, check func(c *gin.Context) error) gin.HandlerFunc {
  return func(c *gin.Context) {
    finished := false
    start := c.GetTime(middleware.StartKey)
    if (start == time.Time{}) {
      start = time.Now()
    }

    go func() {
      err := check(c)
      if !finished {
        finished = true
        if err != nil {
          c.AbortWithError(500, err)
        } else {
          c.Status(200)
          c.Writer.Write([]byte("ok"))
        }
      }
    }()

    time.Sleep(start.Add(timeout).Sub(time.Now()))
    if !finished {
      finished = true
      err := fmt.Errorf("Request exceeded %v.", timeout)
      c.Error(err)
      c.AbortWithStatusJSON(500, interface{}(Error{Err: err.Error()}))
    }
  }
}
