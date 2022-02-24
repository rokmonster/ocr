package main

import (
	"crypto/tls"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	config "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config/serverconfig"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/webcontrollers"
	"github.com/xor22h/rok-monster-ocr-golang/web"
	"golang.org/x/crypto/acme/autocert"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	bolt "go.etcd.io/bbolt"
)

var flags = config.Parse()

func main() {
	rokocr.Prepare(flags.CommonConfiguration)

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
	router.MaxMultipartMemory = 64 << 20
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

	if flags.TLS && len(flags.TLSDomain) > 0 {
		log.Infof("Starting Autocert mode on TLS: https://%v", flags.TLSDomain)
		m := autocert.Manager{
			ForceRSA:   true,
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(flags.TLSDomain),
			Cache:      autocert.DirCache("./.cache"),
		}
		log.Fatal(runWithAutocertManager(router, &m))
	} else {
		log.Infof("Starting in plain HTTP on port: %v", flags.ListenPort)
		log.Fatal(router.Run(fmt.Sprintf(":%d", flags.ListenPort)))
	}
}

func runWithAutocertManager(r http.Handler, m *autocert.Manager) error {
	config := m.TLSConfig()
	config.MinVersion = tls.VersionTLS12

	s := &http.Server{
		Addr:      ":https",
		TLSConfig: config,
		Handler:   r,
	}

	go http.ListenAndServe(":http", m.HTTPHandler(http.HandlerFunc(redirect)))
	return s.ListenAndServeTLS("", "")
}

func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
