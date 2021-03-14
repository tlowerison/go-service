package ratelimiter

import (
  "context"
  "fmt"
  "os"
  "strconv"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/go-redis/redis/v8"
  "github.com/go-redis/redis_rate/v9"
  flag "github.com/spf13/pflag"
  go_service "github.com/tlowerison/go-service"
)

type Middleware struct {
  RateLimit     string
  RedisClient   *redis.Client
  RedisDB       int
  RedisHost     string
  RedisPassword string
  RedisPort     int
  RedisPrefix   string
}

func New() *Middleware {
  return &Middleware{}
}

func (m *Middleware) Register() {
  m.registerEnv()
  m.registerFlags()
}

func (m *Middleware) Handler() gin.HandlerFunc {
  ctx := context.Background()

  if m.RedisClient == nil {
    m.RedisClient = redis.NewClient(&redis.Options{
      Addr:     fmt.Sprintf("%s:%d", m.RedisHost, m.RedisPort),
      Password: m.RedisPassword,
      DB:       m.RedisDB,
    })
  }

  err := m.RedisClient.FlushDB(ctx).Err()
  if err != nil {
    panic(err)
  }

  limiter := redis_rate.NewLimiter(m.RedisClient)
  limit := m.parseRateLimit()

  return func(c *gin.Context) {
    go_service.SetStart(c)
  	res, err := limiter.Allow(ctx, m.RedisPrefix, limit)
  	if err != nil {
  		panic(err)
  	}
    if res.Allowed == 0 {
      c.AbortWithStatus(429)
    } else {
      c.Next()
    }
  }
}

func (m *Middleware) SetRedisClient(redisClient *redis.Client) {
  m.RedisClient = redisClient
}

func (m *Middleware) registerEnv() {
  m.RedisPassword = os.Getenv("REDIS_PASSWORD")
}

func (m *Middleware) registerFlags() {
  flag.StringVar(&m.RateLimit, "rate-limit", "20/m", "Rate limit formatted as [1-9][0-9]+/[hms].")
  flag.IntVar(&m.RedisDB, "redis-db", 0, "DB option for Redis client.")
  flag.StringVar(&m.RedisHost, "redis-host", "localhost", "Host option for Redis client.")
  flag.IntVar(&m.RedisPort, "redis-port", 6379, "Port option for Redis client.")
  flag.StringVar(&m.RedisPrefix, "redis-prefix", "helm_charts_rate_limit", "Key prefix for rate limit caching in Redis.")
}

func (m *Middleware) parseRateLimit() redis_rate.Limit {
  var limit redis_rate.Limit

  components := strings.Split(m.RateLimit, "/")
  if len(components) != 2 {
    panic(fmt.Errorf("Improperly formatted rate limit flag: must follow format [1-9][0-9]+/[hms]"))
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
    limit = redis_rate.PerHour(value)
    break
  case "m":
    limit = redis_rate.PerMinute(value)
    break
  case "s":
    limit = redis_rate.PerSecond(value)
    break
  default:
    panic(fmt.Errorf("Improperly formatted rate limit flag: Must provide rate period as one of the three options: {h,m,s}"))
  }

  return limit
}
