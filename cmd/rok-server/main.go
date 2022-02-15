package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/webcontrollers"
	"github.com/xor22h/rok-monster-ocr-golang/web"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	bolt "go.etcd.io/bbolt"
)

var flags = config.Parse()

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	db, err := bolt.Open("db.bolt", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.New()
	router.SetTrustedProxies([]string{})

	// just reuse same logger
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: log.New().Writer(),
		Formatter: func(params gin.LogFormatterParams) string {
			return fmt.Sprintf("| %3d | %13v | %15s | %-7s %#v\n",
				params.StatusCode,
				params.Latency,
				params.ClientIP,
				params.Method,
				params.Path,
			)
		},
	}))

	router.Use(gin.Recovery())

	router.Use(static.Serve("/", web.EmbeddedFS(web.StaticFS, "static")))
	router.SetHTMLTemplate(web.CreateTemplateEngine(web.StaticFS, "template"))

	pprof.RouteRegister(router.Group("_debug"), "pprof")

	// public group, not auth needed for this.
	public := router.Group("")
	{
		webcontrollers.NewJobsController(public.Group("/jobs"), db).Setup()
		webcontrollers.NewTemplatesController(public.Group("/templates"), flags.TemplatesDirectory, flags.TessdataDirectory).Setup()
	}

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(307, "/jobs")
	})

	log.Fatal(router.Run(":8080"))
}
