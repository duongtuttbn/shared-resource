package ginfx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"tla-backend/pkg/go-kit/log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

const DefaultPort = 8080

// nolint:gochecknoinits
// This block should run in init, so user can override RegisterTagNameFunc if needed without waiting
// for gin.Engine setup completed.
func init() {
	// Support JSON field name when using validator.FieldError.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" {
				name = fld.Name
			}
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// ModuleConfig for the server module.
type ModuleConfig struct {
	Options []fx.Option
}

// ModuleOption apply configuration for ModuleConfig.
type ModuleOption func(*ModuleConfig)

// NewGinModule create a sever module, register all provided Route, Middleware and RouteGroup.
func NewGinModule(options ...ModuleOption) fx.Option {
	conf := ModuleConfig{}
	for _, opt := range options {
		opt(&conf)
	}
	return fx.Module("ginfx",
		fx.Provide(
			NewGinServer,
			NewGinEngine,
		),
		fx.Options(conf.Options...),
		fx.Invoke(func(*http.Server) {}),
	)
}

type GinServerParams struct {
	fx.In
	Engine       *gin.Engine
	ServerConfig ServerConfig
	Lifecycle    fx.Lifecycle
	Shutdowner   fx.Shutdowner
}

// NewGinServer create a http server, backed by gin engine with support for Route, Middleware and RouteGroup.
func NewGinServer(p GinServerParams) (*http.Server, error) {
	port := DefaultPort
	if p.ServerConfig.Port > 0 {
		port = p.ServerConfig.Port
	}
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: p.Engine}
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Infof("Starting HTTP server: %s", srv.Addr)
			go func() {
				err = srv.Serve(ln)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Errorf("Cannot start HTTP server: %v", err)
					_ = p.Shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv, nil
}

type GinEngineParams struct {
	fx.In
	Config       ServerConfig
	Middlewares  []Middleware  `group:"middlewares"`
	RouteGroups  []RouteGroup  `group:"route_groups"`
	RouteModules []RouteModule `group:"route_modules"`
}

// NewGinEngine create a gin engine and register provided Route, RouteGroup, Middleware.
// Also configure log and default middlewares.
func NewGinEngine(p GinEngineParams) (*gin.Engine, error) {
	engine := gin.New()
	if p.Config.TrustedPlatform != "" {
		engine.TrustedPlatform = p.Config.TrustedPlatform
	}

	engine.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, err any) {
		WriteStatus(c, 500)
		if s, ok := err.(string); ok {
			log.Errorf("panic!: " + s)
			return
		}
		log.Errorf("panic!: %+v", err)
	}))

	if len(p.Config.CorsAllowedOrigins) > 0 {
		engine.Use(newCORSMiddleware(p.Config))
	}
	if p.Config.ServerTimeout > 0 {
		engine.Use(newTimeoutMiddleware(p.Config))
	}

	slices.SortStableFunc(p.Middlewares, func(a, b Middleware) int {
		return a.Priority() - b.Priority()
	})

	// Apply global middlewares
	for _, middleware := range p.Middlewares {
		engine.Use(middleware.Handle)
	}

	ginGroupByName, err := buildRouteGroupMap(engine, p.RouteGroups)
	if err != nil {
		return nil, err
	}

	for _, routeModule := range p.RouteModules {
		if r, ok := routeModule.(RouteModuleWithGroup); ok && r.Group() != "" {
			group := r.Group()
			ginGroup, found := ginGroupByName[group]
			if !found {
				return nil, fmt.Errorf("route group %s not found", group)
			}

			routeModule.Router(ginGroup)
			continue
		}

		routeModule.Router(engine)
	}

	metric := NewPrometheusMetric()
	metric.Use(engine)

	return engine, nil
}

type routeGroupNode struct {
	Group    RouteGroup
	Children []*routeGroupNode
}

func buildRouteGroupMap(engine gin.IRouter, groups []RouteGroup) (map[RouteGroupName]gin.IRouter, error) {
	groupMap := make(map[RouteGroupName]*routeGroupNode)
	var roots []*routeGroupNode

	// Create nodes
	for _, g := range groups {
		groupMap[g.Name()] = &routeGroupNode{Group: g}
	}

	// Build tree
	for _, node := range groupMap {
		parentName := node.Group.ParentGroup()
		if parentName == "" {
			roots = append(roots, node)
			continue
		}
		if parent, ok := groupMap[parentName]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}

		return nil, errors.Errorf("route group %s not found", parentName)
	}

	results := make(map[RouteGroupName]gin.IRouter, len(groups))

	for _, root := range roots {
		registerGroups(engine, root, results)
	}

	return results, nil
}

func registerGroups(r gin.IRouter, node *routeGroupNode, groupMap map[RouteGroupName]gin.IRouter) {
	g := node.Group.Route(r)
	groupMap[node.Group.Name()] = g
	for _, child := range node.Children {
		registerGroups(g, child, groupMap)
	}
}
