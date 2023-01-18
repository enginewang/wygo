package middleware

import (
	"github.com/enginewang/wygo"
	"log"
	"time"
)

func Logger() wygo.HandlerFunc {
	return func(c *wygo.Context) error {
		t := time.Now()
		// 先调用里面的，计算的是包含在内的所有的运行时间
		c.Next()
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
		return nil
	}
}
