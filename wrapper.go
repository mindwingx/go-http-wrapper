package httpwrapper

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/mindwingx/abstraction"
	"github.com/mindwingx/go-helper"
	"time"
)

// INSTANTIATE

type (
	engine struct {
		config httpConfig
		locale abstraction.Locale
		core   *gin.Engine
		ctx    *gin.Context
		group  *gin.RouterGroup
	}

	httpConfig struct {
		Host            string
		Port            string
		Development     bool
		ShutdownTimeout time.Duration
	}
)

func NewGin(registry abstraction.Registry, locale abstraction.Locale) abstraction.ApiService {
	http := new(engine)
	err := registry.Parse(&http.config)
	if err != nil {
		helper.CustomPanic("", err)
	}

	http.locale = locale
	http.core = gin.New()

	if http.config.Development == false {
		gin.SetMode(gin.ReleaseMode)
	}

	http.ctx = &gin.Context{}

	return http
}

func (g *engine) InitApiService() {
	//todo: handle limitation of calling the current service.
	//for docker, use the other services names' on the same network
	err := g.core.SetTrustedProxies([]string{"0.0.0.0"})
	if err != nil {
		helper.CustomPanic(g.locale.Get("http_init_err"), err)
	}
	g.core.Use(gin.Logger())
	g.core.Use(gin.Recovery())
}

func (g *engine) StartHttp() {
	address := fmt.Sprintf("%s:%s", g.config.Host, g.config.Port)
	color.Blue(g.locale.Get("http_start"))

	//todo: handle running the service by http.ListenAndServe for shutdown and restart.
	err := g.core.Run(address)
	if err != nil {
		helper.CustomPanic(g.locale.Get("http_serve_err"), err)
	}
}

// WRAPPERS

// the below struct naming convention will use to wrap any framework

func (g *engine) Param(key string) string {
	return g.ctx.Param(key)
}

func (g *engine) Query(key string) string {
	return g.ctx.Query(key)
}

func (g *engine) GetHeader(key string) string {
	return g.ctx.GetHeader(key)
}

func (g *engine) BindJSON(obj interface{}) (err error) {
	err = g.ctx.BindJSON(obj)

	if err != nil {
		return err
	}

	return nil
}

func (g *engine) ShouldBindJSON(obj interface{}) (err error) {
	err = g.ctx.ShouldBindJSON(obj)

	if err != nil {
		return err
	}

	return nil
}

func (g *engine) JSON(code int, obj interface{}) {
	g.ctx.JSON(code, obj)
}

func (g *engine) AbortWithStatusJSON(status int, data interface{}) {
	g.ctx.AbortWithStatusJSON(status, data)
}

func (g *engine) Abort() {
	g.ctx.Abort()
}

func (g *engine) Next() {
	g.ctx.Next()
}

// route group

func (g *engine) RouteGroup(prefix string) abstraction.RouteGroup {
	g.group = g.core.Group(prefix)
	return g
}

func (g *engine) NestedGroup(prefix string) abstraction.RouteGroup {
	g.group = g.group.Group(prefix)
	return g
}

// wrapping HTTP verbs

func (g *engine) Get(path string, handler ...abstraction.CustomCtxFunc) {
	g.router().GET(path, g.handleContext(handler...)...)
}

func (g *engine) Post(path string, handler ...abstraction.CustomCtxFunc) {
	g.router().POST(path, g.handleContext(handler...)...)
}

func (g *engine) Put(path string, handler ...abstraction.CustomCtxFunc) {
	g.router().PUT(path, g.handleContext(handler...)...)
}

func (g *engine) Delete(path string, handler ...abstraction.CustomCtxFunc) {
	g.router().DELETE(path, g.handleContext(handler...)...)
}

// HELPER METHODS

func (g *engine) router() gin.IRoutes {
	var router gin.IRoutes

	if g.group != nil {
		router = g.group
	} else {
		router = g.core
	}

	return router
}

func (g *engine) handleContext(handlers ...abstraction.CustomCtxFunc) []gin.HandlerFunc {
	var handlerFunctions []gin.HandlerFunc

	for _, handler := range handlers {
		handlerFunctions = append(handlerFunctions, g.getHandler(handler))
	}

	return handlerFunctions
}

func (g *engine) getHandler(handler abstraction.CustomCtxFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c)
	}
}
