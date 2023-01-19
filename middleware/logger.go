package middleware

import (
	"log"
	"time"
	"wygo"
)

func Logger() wygo.HandlerFunc {
	return func(c *wygo.Context) {
		t := time.Now()
		// 先调用里面的，计算的是包含在内的所有的运行时间
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
