package wygo

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		// 对于每种类型的方法，存各个方法的node根节点
		roots: make(map[string]*node),
		// 对于各个具体的pattern，存对应的handlerFunc
		handlers: make(map[string]HandlerFunc),
	}
}

// 将pattern解析到parts数组中
// 如果遇到*的话，只看第一个*并且直接返回
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 添加一个router
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	// 往对应方法的Trie中插入一个节点
	r.roots[method].insert(pattern, parts, 0)
	// 另外存一下handler
	r.handlers[key] = handler
}

// 获取router，返回对应的node和param map
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	// 先搜索匹配这个parts的的所有node
	n := root.search(searchParts, 0)
	// 如果搜索到了
	if n != nil {
		// 解析这个pattern字符串
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			// 如果part开头是:，就将搜索到的子节点作为对应的param进行存储
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			// 如果是*的话，就将剩下的都作为param
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
