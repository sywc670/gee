package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(ctx *Context) {
		t := time.Now()

		ctx.Next()

		log.Printf("[%d] URI: %s TIME: %v", ctx.StatusCode, ctx.Request.RequestURI, time.Since(t))
	}
}
