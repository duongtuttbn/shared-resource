package ginfx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type RouteGroupName string

// RouteGroup interface provide integration of gin.RouterGroup into fx.
type RouteGroup interface {
	ParentGroup() RouteGroupName
	// Route apply routing logic using group gin.IRouter.
	// return the created group so the framework can apply middleware to it.
	Route(r gin.IRouter) gin.IRouter
	// Name return the RouteGroupName of this RouteGroup.
	// Name must not empty
	Name() RouteGroupName
}

type RouteGroupHandler struct {
	name        RouteGroupName
	parentGroup RouteGroupName
	prefix      string
	handlers    []gin.HandlerFunc
}

var _ RouteGroup = RouteGroupHandler{}

func NewRouteGroup(name RouteGroupName, prefix string, handlers ...gin.HandlerFunc) RouteGroupHandler {
	return RouteGroupHandler{
		name:     name,
		prefix:   prefix,
		handlers: handlers,
	}
}

func (g RouteGroupHandler) Route(r gin.IRouter) gin.IRouter {
	return r.Group(g.prefix, g.handlers...)
}

func (g RouteGroupHandler) Name() RouteGroupName {
	return g.name
}

func (g RouteGroupHandler) ParentGroup() RouteGroupName {
	return g.parentGroup
}

func (g RouteGroupHandler) WithParent(parentName RouteGroupName) RouteGroupHandler {
	g.parentGroup = parentName
	return g
}

// AsGroup register function into route_groups.
func AsGroup(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(RouteGroup)),
		fx.ResultTags(`group:"route_groups"`),
	)
}

type RouteModule interface {
	Router(r gin.IRouter)
}

type RouteModuleWithGroup interface {
	Group() RouteGroupName
}

// AsModule register function into route_modules.
func AsModule(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(RouteModule)),
		fx.ResultTags(`group:"route_modules"`),
	)
}
