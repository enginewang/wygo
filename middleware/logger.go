package middleware

import (
	"github.com/enginewang/wygo"
	"github.com/enginewang/wygo/log"
	"time"
)

func Logger() wygo.HandlerFunc {
	return func(c *wygo.Context) {
		t := time.Now()
		//fmt.Println(c.StatusCode)
		// 先调用里面的，计算的是包含在内的所有的运行时间
		c.Next()
		//fmt.Println(c.StatusCode)
		if c.StatusCode != 0 {
			log.Infof("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
		}
	}
}
