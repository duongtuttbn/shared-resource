package ginswagger

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

//go:embed index.html
var indexHTML string

type Params struct {
	fx.In
	Engine *gin.Engine
	Config SwaggerConfig `optional:"true"`
	Data   Data          `optional:"true"`
}

func NewModule() fx.Option {
	return fx.Module("ginswagger",
		fx.Invoke(func(params Params) {
			cfg := params.Config
			if cfg.SwaggerAPIHost == "" || params.Data.JSON == "" {
				return
			}

			docJSON := strings.Replace(params.Data.JSON, "\"swagger\": \"2.0\"", "\"swagger\": \"2.0\",\"host\": \""+cfg.SwaggerAPIHost+"\"", 1)

			var handlers []gin.HandlerFunc
			if cfg.SwaggerBasicAuthPassword != "" {
				handlers = append(handlers, gin.BasicAuthForRealm(map[string]string{
					lo.CoalesceOrEmpty(cfg.SwaggerBasicAuthUser, "admin"): cfg.SwaggerBasicAuthPassword,
				}, "Enter password"))
			}

			params.Engine.Group(fmt.Sprintf("%s/swagger", cfg.SwaggerHiddenPath), handlers...).
				GET("/index.html", func(c *gin.Context) {
					c.Header("Content-Type", "text/html; charset=utf-8")
					c.String(http.StatusOK, indexHTML)
				}).
				GET("/doc.json", func(c *gin.Context) {
					c.Header("Content-Type", "application/json")
					c.String(http.StatusOK, docJSON)
				})
		}),
	)
}
