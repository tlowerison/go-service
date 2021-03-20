package health

import (
  "fmt"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/rs/zerolog/log"
  flag "github.com/spf13/pflag"
  go_service "github.com/tlowerison/go-service"
)

type Service struct {
  Check   Check
  Port    int
  Timeout string
  timeout time.Duration
}

type Error struct {
  Err string `json:"error"`
}

type Check func(c *gin.Context) error

func New(check Check) *Service {
  return &Service{
    Check: check,
  }
}

func (s *Service) Register() {
  flag.IntVar(&s.Port, "health-port", 2020, "Health port to listen on.")
  flag.StringVar(&s.Timeout, "health-timeout", "10-s", "Rate limit formatted as [1-9][0-9]+-[hms].")
}

func (s *Service) Serve() <-chan *http.Server {
  channel := make(chan *http.Server)
  go func() {
    defer close(channel)

    router := gin.New()
    server := &http.Server{Addr: fmt.Sprintf(":%d", s.Port), Handler: router}

    router.GET("/health", s.handler())

  	log.Info().Msgf("Health server starting on port %d.", s.Port)

    channel <- server

    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
  		log.Fatal().Err(err).Msg("Starting health server failed.")
  	}
  }()
  return channel
}

func (s *Service) handler() gin.HandlerFunc {
  s.parseTimeout()
  return func(c *gin.Context) {
    start := c.GetTime(go_service.StartKey)
    if (start == time.Time{}) {
      start = time.Now()
    }

    var err error
    select {
    case err = <-waitForCheck(c, s.Check):
    case err = <-waitForTimeout(start, s.timeout):
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

func (s *Service) parseTimeout() {
  components := strings.Split(s.Timeout, "-")
  if len(components) != 2 {
    panic(fmt.Errorf("Improperly formatted rate limit flag: must follow format [1-9][0-9]+-[hms]"))
  }

  value, err := strconv.Atoi(components[0])
  if err != nil {
    panic(fmt.Errorf("Improperly formatted rate limit flag: %s", err.Error()))
  }

  if value < 0 {
    panic(fmt.Errorf("Improperly formatted rate limit flag: Cannot use negative limits: %d", value))
  }

  period := components[1]

  switch period {
  case "h":
    s.timeout = time.Duration(time.Duration(value) * time.Hour)
    break
  case "m":
    s.timeout = time.Duration(time.Duration(value) * time.Minute)
    break
  case "s":
    s.timeout = time.Duration(time.Duration(value) * time.Second)
    break
  default:
    panic(fmt.Errorf("Improperly formatted rate limit flag: Must provide rate period as one of the three options: {h,m,s}"))
  }
}
