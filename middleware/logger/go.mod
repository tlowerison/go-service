module github.com/tlowerison/go-service/logger

go 1.16

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/rs/zerolog v1.20.0
	github.com/tlowerison/go-service/middleware v0.0.0-00010101000000-000000000000
// github.com/tlowerison/go-service/middleware v0.0.0-20210314101422-11a7269550ba
)

replace github.com/tlowerison/go-service/middleware => ../middleware
