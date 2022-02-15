package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/webcontrollers"
	"github.com/xor22h/rok-monster-ocr-golang/web"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	router.SetTrustedProxies([]string{})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(static.Serve("/", web.EmbeddedFS(web.StaticFS, "static")))
	router.SetHTMLTemplate(web.CreateTemplateEngine(web.StaticFS, "template"))

	pprof.RouteRegister(router.Group("_debug"), "pprof")

	// public group, not auth needed for this.
	public := router.Group("")
	{
		webcontrollers.NewJobsController(public.Group("/jobs")).Setup()
		webcontrollers.NewTemplatesController(public.Group("/templates")).Setup()
	}

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(307, "/jobs")
	})

	log.Fatal(router.Run(":8080"))
}
