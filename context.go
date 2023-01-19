package wygo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type J map[string]interface{}

type Context struct {
	// 原生的ResponseWriter和Request
	Writer http.ResponseWriter
	Req    *http.Request
	// 封装request的context
	Ctx context.Context
	// Request的一些常用字段的直接访问
	mu     *sync.Mutex
	Method string
	Path   string
	// 存储的params数组，提供对路由参数的访问
	Params map[string]string
	// Response的一些常用字段的直接访问
	StatusCode int
	// 中间件
	handlers     []HandlerFunc
	handlerIndex int
	// engine pointer
	engine *Engine
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:       w,
		Req:          r,
		Ctx:          r.Context(),
		Path:         r.URL.Path,
		Method:       r.Method,
		handlerIndex: -1,
	}
}

func (c *Context) GetHandlers() []HandlerFunc {
	return c.handlers
}

func (c *Context) SetHandlers(handlers []HandlerFunc) {
	c.handlers = handlers
}

func (c *Context) Next() {
	c.handlerIndex++
	s := len(c.handlers)
	for ; c.handlerIndex < s; c.handlerIndex++ {
		c.handlers[c.handlerIndex](c)
	}
}

//func (c *Context) Next() error {
//	if c.handlerIndex < len(c.handlers) {
//		if err := c.handlers[c.handlerIndex](c); err != nil {
//			return err
//		}
//	}
//	c.handlerIndex += 1
//	return nil
//}

// 最基本的对各种属性的Get/Set

func (c *Context) Mux() *sync.Mutex {
	return c.mu
}

func (c *Context) GetRequest() *http.Request {
	return c.Req
}

func (c *Context) SetRequest(r *http.Request) {
	c.Req = r
}

func (c *Context) GetRespnse() http.ResponseWriter {
	return c.Writer
}

func (c *Context) SetResponse(w http.ResponseWriter) {
	c.Writer = w
}

func (c *Context) BaseContext() context.Context {
	return c.Ctx
}

// 实现基础context的所有接口方法，从而与其他第三方库的context可以联动
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.BaseContext().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.BaseContext().Done()
}

func (c *Context) Err() error {
	return c.BaseContext().Err()
}

func (c *Context) Value(key any) any {
	return c.BaseContext().Value(key)
}

// 一些便捷方法的封装，包括Req和Resp
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 一些关于Request的封装
// 这里调用net/url的Query()方法，对url的?后面的query进行解析
// 比如 xxx:port/get?a=1&a=2&b=3会被解析为 {a:[1,2], b:[3]}
// 对于a来说，最后出现的2是有效的，之前的会被替代

func (c *Context) QueryAll() map[string][]string {
	if c.Req != nil {
		return c.Req.URL.Query()
	}
	return map[string][]string{}
}

// 查询key对应的int值，如果没找到就用def(ault)代替

func (c *Context) QueryInt(key string, defaultValue int) int {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		val := vals[len(vals)-1]
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return defaultValue
		}
		return valInt
	}
	return defaultValue
}

func (c *Context) QueryFloat64(key string, defaultValue float64) float64 {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		val := vals[len(vals)-1]
		valInt, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return defaultValue
		}
		return valInt
	}
	return defaultValue
}

func (c *Context) QueryFloat32(key string, defaultValue float64) float64 {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		val := vals[len(vals)-1]
		valInt, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return defaultValue
		}
		return valInt
	}
	return defaultValue
}

func (c *Context) QueryString(key string, defaultValue string) string {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		return vals[len(vals)-1]
	}
	return defaultValue
}

func (c *Context) QueryArray(key string, defaultValue []string) []string {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		return vals
	}
	return defaultValue
}

func (c *Context) PostFormInt(key string, defaultValue int) int {
	valInt, err := strconv.Atoi(c.Req.FormValue(key))
	if err != nil {
		return defaultValue
	}
	return valInt
}

func (c *Context) PostFormFloat64(key string, defaultValue float64) float64 {
	val, err := strconv.ParseFloat(c.Req.FormValue(key), 64)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Context) PostFormFloat32(key string, defaultValue float64) float64 {
	val, err := strconv.ParseFloat(c.Req.FormValue(key), 32)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Context) SetStatusCode(statusCode int) {
	c.StatusCode = statusCode
	c.Writer.WriteHeader(statusCode)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Fail(statusCode int, msg string) {
	c.SetStatusCode(statusCode)
	log.Println("Error:" + msg)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// 一些便捷返回类型的封装，包括String,HTML,Data,JSON
func (c *Context) Bytes(statusCode int, b []byte) error {
	c.SetStatusCode(statusCode)
	_, err := c.Writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) String(statusCode int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.SetStatusCode(statusCode)
	_, err := c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
	if err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
	}
}

func (c *Context) JSON(statusCode int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.SetStatusCode(statusCode)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
	}
}

func (c *Context) HTML(statusCode int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.SetStatusCode(statusCode)
	_, err := c.Writer.Write([]byte(html))
	if err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
	}
}

func (c *Context) HTMLTemplate(statusCode int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.SetStatusCode(statusCode)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
