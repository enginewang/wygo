package middleware

import (
	"github.com/enginewang/wlog"
	"github.com/enginewang/wygo"
)

func M1() wygo.HandlerFunc {
	return func(c *wygo.Context) {
		wlog.Infof("[M1 Begin] %v", c.Req.RequestURI)
		c.Next()
		wlog.Infof("[M1 End] %v", c.Req.RequestURI)
	}
}

func M2() wygo.HandlerFunc {
	return func(c *wygo.Context) {
		wlog.Infof("[M2 Begin] %v", c.Req.RequestURI)
		c.Next()
		wlog.Infof("[M2 End] %v", c.Req.RequestURI)
	}
}
