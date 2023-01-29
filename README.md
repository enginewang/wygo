# wygo

A Simple Go Web Framework

借鉴了gin、echo、 martini、gee等框架的一些设计思路

- 路由Trie树
- Context封装
- 支持自定义中间件
- 支持JSON等多种返回格式，支持HTML模板

自带一些常用中间件，可自行配置
- Logger
- Recover
- Basic Auth（待完成）
- JWT（待完成）
- CORS（待完成）


特性：
- 支持Response返回值链式调用
- 支持中间件链式调用，支持单个中间件或者中间件链
- 支持自定义Log等级，彩色显示多种类型的Log信息

待完成：
- 支持Dockerfile和Docker-compose打包
