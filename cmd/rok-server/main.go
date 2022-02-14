package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
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
		public.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{})
		})
	}

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "index.html", gin.H{})
	})

	log.Fatal(router.Run(":8080"))
}
