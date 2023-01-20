package wygo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type IRequest interface {
	// Query相关，也就是url &a=1这样的
	Query(key string) interface{}
	QueryInt(key string, defaultValue int) int
	QueryInt64(key string, defaultValue int64) int64
	QueryFloat64(key string, defaultValue float64) float64
	QueryFloat32(key string, defaultValue float32) float32
	//QueryBool(key string, defaultValue bool) (bool, bool)
	QueryString(key string, defaultValue string) string
	QueryStringSlice(key string, defaultValue []string) []string
	// Param相关，也就是/:id这样的
	Param(key string) interface{}
	ParamInt(key string, defaultValue int) int
	ParamInt64(key string, defaultValue int64) int64
	ParamFloat64(key string, defaultValue float64) float64
	ParamFloat32(key string, defaultValue float32) float32
	//ParamBool(key string, defaultValue bool) (bool, bool)
	ParamString(key string, defaultValue string) string
	// Form表单数据的获取
	PostForm(key string) string
	PostFormInt(key string, defaultValue int) int
	PostFormInt64(key string, defaultValue int64) int64
	PostFormFloat64(key string, defaultValue float64) float64
	PostFormFloat32(key string, defaultValue float32) float32
	//PostFormBool(key string, defaultValue bool) (bool, bool)
	//PostFormString(key string) string

	// 将body文本解析到obj中
	BindJson(obj interface{}) error

	//PostFormStringSlice(key string, defaultValue []string) (string, bool)
	//PostFormFile(key string) (*multipart.FileHeader, error)
	//BindXML(obj interface{}) error

	//Uri() string
	//Method() string
	//Host() string
	//ClientIP() string

	//Headers() map[string]string
	//Header(key string) (string, bool)
	//
	//Cookies() map[string]string
	//Cookie(key string) (string, bool)
}

type IResponse interface {
	// 返回的仍然是IResponse，允许方法链式调用，提高可读性
	// 返回类型
	JSON(obj interface{}) IResponse
	HTML(html string) IResponse
	HTMLTemplate(template string, data interface{}) IResponse
	String(format string, values ...interface{}) IResponse
	// 重定向
	Redirect(path string) IResponse
	SetHeader(key string, val string) IResponse
	SetStatusCode(code int) IResponse
	SetCookie(key string, val string, maxAge int, path, domain string, secure, httpOnly bool) IResponse
	SetStatusOK() IResponse
	SetStatusInternalServerError() IResponse
}

// 一些关于Request的封装
// 这里调用net/url的Query()方法，对url的?后面的query进行解析
// 比如 xxx:port/get?a=1&a=2&b=3会被解析为 {a:[1,2], b:[3]}
// 对于a来说，最后出现的2是有效的，之前的会被替代
func (c *Context) Query(key string) interface{} {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) QueryAll() map[string][]string {
	if c.Req != nil {
		return map[string][]string(c.Req.URL.Query())
	}
	return map[string][]string{}
}

func (c *Context) QueryInt(key string, defaultValue int) int {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		if len(vals) > 0 {
			valInt, err := strconv.Atoi(vals[len(vals)-1])
			if err != nil {
				return defaultValue
			}
			return valInt
		}
	}
	return defaultValue
}

func (c *Context) QueryInt64(key string, defaultValue int64) int64 {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		if len(vals) > 0 {
			valInt64, err := strconv.Atoi(vals[len(vals)-1])
			if err != nil {
				return defaultValue
			}
			return int64(valInt64)
		}
	}
	return defaultValue
}

func (c *Context) QueryFloat64(key string, defaultValue float64) float64 {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		if len(vals) > 0 {
			valFloat64, err := strconv.ParseFloat(vals[len(vals)-1], 64)
			if err != nil {
				return defaultValue
			}
			return valFloat64
		}
	}
	return defaultValue
}

func (c *Context) QueryFloat32(key string, defaultValue float32) float32 {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		if len(vals) > 0 {
			valFloat32, err := strconv.ParseFloat(vals[len(vals)-1], 32)
			if err != nil {
				return defaultValue
			}
			return float32(valFloat32)
		}
	}
	return defaultValue
}

func (c *Context) QueryString(key string, defaultValue string) string {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		if len(vals) > 0 {
			return vals[len(vals)-1]
		}
	}
	return defaultValue
}

func (c *Context) QueryStringSlice(key string, defaultValue []string) []string {
	params := c.QueryAll()
	if vals, ok := params[key]; ok {
		return vals
	}
	return defaultValue
}

func (c *Context) Param(key string) interface{} {
	value, _ := c.Params[key]
	return value
}

func (c *Context) ParamInt(key string, defaultValue int) int {
	if value, ok := c.Params[key]; ok {
		valInt, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return valInt
	}
	return defaultValue
}

func (c *Context) ParamInt64(key string, defaultValue int64) int64 {
	if value, ok := c.Params[key]; ok {
		valInt64, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return int64(valInt64)
	}
	return defaultValue
}

func (c *Context) ParamFloat64(key string, defaultValue float64) float64 {
	if value, ok := c.Params[key]; ok {
		valFloat64, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultValue
		}
		return valFloat64
	}
	return defaultValue
}

func (c *Context) ParamFloat32(key string, defaultValue float32) float32 {
	if value, ok := c.Params[key]; ok {
		valFloat32, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return defaultValue
		}
		return float32(valFloat32)
	}
	return defaultValue
}

func (c *Context) ParamString(key string, defaultValue string) string {
	if value, ok := c.Params[key]; ok {
		return value
	}
	return defaultValue
}

// 一些便捷方法的封装，包括Req和Resp
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) PostFormInt(key string, defaultValue int) int {
	valInt, err := strconv.Atoi(c.PostForm(key))
	if err != nil {
		return defaultValue
	}
	return valInt
}

func (c *Context) PostFormInt64(key string, defaultValue int64) int64 {
	valInt64, err := strconv.Atoi(c.PostForm(key))
	if err != nil {
		return defaultValue
	}
	return int64(valInt64)
}

func (c *Context) PostFormFloat64(key string, defaultValue float64) float64 {
	valFloat64, err := strconv.ParseFloat(c.PostForm(key), 64)
	if err != nil {
		return defaultValue
	}
	return valFloat64
}

func (c *Context) PostFormFloat32(key string, defaultValue float32) float32 {
	valFloat32, err := strconv.ParseFloat(c.PostForm(key), 32)
	if err != nil {
		return defaultValue
	}
	return float32(valFloat32)
}

func (c *Context) BindJson(obj interface{}) error {
	if c.Req != nil {
		body, err := io.ReadAll(c.Req.Body)
		if err != nil {
			return err
		}
		// 因为Body只能读一次，后面会读不到，所以再构建一个赋值给之前的Body
		c.Req.Body = io.NopCloser(bytes.NewBuffer(body))
		err = json.Unmarshal(body, obj)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Request Empty")
	}
	return nil
}

func (c *Context) SetStatusCode(statusCode int) IResponse {
	c.Writer.WriteHeader(statusCode)
	c.StatusCode = statusCode
	return c
}

func (c *Context) SetCookie(key string, val string, maxAge int, path, domain string, secure, httpOnly bool) IResponse {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     key,
		Value:    url.QueryEscape(val),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: 1,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
	return c
}

func (c *Context) SetStatusOK() IResponse {
	c.Writer.WriteHeader(http.StatusOK)
	return c
}

func (c *Context) SetStatusInternalServerError() IResponse {
	c.Writer.WriteHeader(http.StatusInternalServerError)
	return c
}

func (c *Context) SetHeader(key string, value string) IResponse {
	c.Writer.Header().Add(key, value)
	return c
}

func (c *Context) Redirect(path string) IResponse {
	http.Redirect(c.Writer, c.Req, path, http.StatusMovedPermanently)
	return c
}

// 一些便捷返回类型的封装，包括String,HTML,Data,JSON

func (c *Context) String(format string, values ...interface{}) IResponse {
	c.SetHeader("Content-Type", "text/plain")
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
	return c
}

func (c *Context) JSON(obj interface{}) IResponse {
	c.SetHeader("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
	}
	return c
}

func (c *Context) HTML(html string) IResponse {
	c.SetHeader("Content-Type", "text/html")
	_, err := c.Writer.Write([]byte(html))
	if err != nil {
		c.SetStatusCode(http.StatusInternalServerError)
	}
	return c
}

func (c *Context) HTMLTemplate(template string, data interface{}) IResponse {
	c.SetHeader("Content-Type", "text/html")
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, template, data); err != nil {
		c.SetStatusInternalServerError()
	}
	return c
}
