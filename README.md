# wygo

A Simple Go Web Framework

借鉴了gin、echo、 martini、gee等框架的一些设计思路

- 路由Trie树
- Context封装
- 支持自定义中间件
- 支持JSON等多种返回格式，支持HTML模板
- 嵌套ORM框架
- 性能出色

自带一些常用中间件：
- Basic Auth
- JWT
- CORS
- CSRF
- GZIP
- Logger
- Recover

特性：
- 返回值链式调用
```go

```

支持中间件链式调用，支持单个中间件或者中间件链

中间件打印，Abort暂停

重复调用相同中间件，会warning

支持viper配置文件

支持Dockerfile和Docker-compose打包

支持Jeckfin

封装彩色Log库wlog，支持自定义Log等级、颜色、log文件输出位置。