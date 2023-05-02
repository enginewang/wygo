package wygo

import (
	"fmt"
	"github.com/enginewang/wygo/log"
	"html/template"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
)

const (
	LOGO = `
 _      ____  ______ _____ 
| | /| / / / / / __ |/ __ \
| |/ |/ / /_/ / /_/ / /_/ /
|__/|__/\__  /\__  /\____/
       /____//____/        
`
)

type Config struct {
	// 一个路由出现相同的多个middleware会报warning提示，默认开启
	SameMiddlewareWarning bool
	// 后台监测页面，默认为true
	WatchSystem bool
	// 设置后台监管页面的账户
	User     string
	Password string
}

type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	engine      *Engine       // all groups share a Engine instance
}

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // store all groups
	// html模板
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
	config        *Config
}

// New is the constructor of wygo.Engine
func New() *Engine {
	engine := &Engine{router: newRouter(), config: newConfig()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	fmt.Println(LOGO)
	return engine
}

func newConfig() *Config {
	return &Config{
		SameMiddlewareWarning: true,
		WatchSystem:           true,
		User:                  "",
		Password:              "",
	}
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Infof("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

func (group *RouterGroup) UPDATE(pattern string, handler HandlerFunc) {
	group.addRoute("UPDATE", pattern, handler)
}

func (group *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}

func (group *RouterGroup) PATCH(pattern string, handler HandlerFunc) {
	group.addRoute("PATCH", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	fmt.Printf("Wygo Serve on %v\n", addr)
	for _, group := range engine.groups {
		if group.prefix != "" {
			group.PrintMiddlewares()
		}
	}
	return http.ListenAndServe(addr, engine)
}

type MiddlewareHandler interface {
	BeUsed(*RouterGroup)
}

func (h HandlerFunc) BeUsed(group *RouterGroup) {
	group.middlewares = append(group.middlewares, h)
}

func (mc MiddlewareChain) BeUsed(group *RouterGroup) {
	group.middlewares = append(group.middlewares, mc.Middlewares...)
}

func (group *RouterGroup) Use(mh ...MiddlewareHandler) *RouterGroup {
	//group.middlewares = append(group.middlewares, middlewares...)
	for _, m := range mh {
		m.BeUsed(group)
	}
	return group
}

func (group *RouterGroup) UseChain(chain *MiddlewareChain) *RouterGroup {
	if chain != nil && len(chain.Middlewares) > 0 {
		group.middlewares = append(group.middlewares, chain.Middlewares...)
	}
	return group
}

func getFuncName(i interface{}) string {
	tmp := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")
	return strings.Join(tmp[len(tmp)-2:len(tmp)-1], ".")
}

func (group *RouterGroup) PrintMiddlewares() {
	str := fmt.Sprintf("Group [%v] \t [Middlewares] \t ", group.prefix)
	for i, middleware := range group.middlewares {
		str += getFuncName(middleware)
		if i != len(group.middlewares)-1 {
			str += "<==>"
		}
	}
	log.Info(str)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.ParamString("filepath", "")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.SetStatusCode(http.StatusNotFound)
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

type MiddlewareChain struct {
	Middlewares []HandlerFunc
}

func (engine *Engine) NewMiddlewareChain(middlewares ...HandlerFunc) *MiddlewareChain {
	var mc []HandlerFunc
	return &MiddlewareChain{Middlewares: append(mc, middlewares...)}
}
