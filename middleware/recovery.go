package middleware

import (
	"fmt"
	"github.com/enginewang/wygo"
	"log"
	"runtime"
	"strings"
)

// 打印堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() wygo.HandlerFunc {
	return func(c *wygo.Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.SetStatusInternalServerError()
			}
		}()
		c.Next()
	}
}
