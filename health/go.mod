module github.com/tlowerison/go-service/health

go 1.16

replace github.com/tlowerison/go-service/middleware => ../middleware

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/tlowerison/go-service/middleware v0.0.0-00010101000000-000000000000
)
