package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rokmonster/ocr/internal/pkg/www"
	"github.com/rokmonster/ocr/internal/pkg/www/middlewares"

	"github.com/rokmonster/ocr/internal/pkg/utils"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	config "github.com/rokmonster/ocr/internal/pkg/config/serverconfig"
	"github.com/rokmonster/ocr/internal/pkg/rokocr"
	"github.com/rokmonster/ocr/web"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/acme/autocert"
)

var flags = config.Parse()

func main() {
	if flags.Install {
		if len(strings.TrimSpace(flags.InstallUser)) > 0 && flags.TLS && len(strings.TrimSpace(flags.TLSDomain)) > 0 {
			fmt.Printf("# Generating systemd unit files for running as %v\n", flags.InstallUser)
			rokocr.InstallSystemD(flags)
			fmt.Println("# Please run commands below")
			fmt.Println("sudo systemctl daemon-reload")
			fmt.Println("sudo systemctl enable --now rokocr-server.service")
			os.Exit(0)
		} else {
			fmt.Println("# Install only possible with tls && domain name")
			os.Exit(1)
		}
	}

	rokocr.Prepare(flags.CommonConfiguration)
	rokocr.DownloadTesseractData(flags.CommonConfiguration)
	rokocr.PreloadTemplates(flags.CommonConfiguration)

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	db, err := bolt.Open("db.bolt", 0666, &bolt.Options{Timeout: time.Second})
	utils.Panic(err)

	defer db.Close()

	router := gin.New()
	_ = router.SetTrustedProxies([]string{})

	// just reuse same logger
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: log.New().WriterLevel(log.DebugLevel),
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

	oauth := middlewares.NewOAuth2Middleware(router, flags.OAuthClientID, flags.OAuthSecretID, flags.TLSDomain)

	rootRouter := router.Group("")
	{
		devices := rootRouter.Group("/devices")
		{
			controller := www.NewRemoteDevicesController(flags.TemplatesDirectory, flags.TessdataDirectory)

			// websocket - no-auth (different auth in future)
			devices.GET("/ws", controller.Websocket)
			// everything else - requires auth-middleware
			devicesSecure := devices.Group("", oauth.Middleware())
			{
				devicesSecure.GET("/", controller.GetListOfDevices)
				devicesSecure.GET("/:id/disconnect", controller.Disconnect)
				devicesSecure.GET("/:id/data", controller.Data)
			}
		}

		api := rootRouter.Group("/api")
		{
			apiController := www.NewAPIController(flags.TessdataDirectory)
			api.POST("/hoh", apiController.ScanHOH)
		}

		// all job's API's require auth
		jobs := rootRouter.Group("/jobs", oauth.Middleware())
		{
			controller := www.NewJobsController(db, flags.TessdataDirectory)
			jobs.GET("/", controller.GetJobsList)
			jobs.GET("/create", controller.CreateJobForm)
			jobs.GET("/:id", controller.GetJobByID)
			jobs.GET("/:id/start", controller.StartJobByID)
			jobs.GET("/:id/csv", controller.ExportJobAsCSV)
			jobs.GET("/:id/results", controller.ExportJobResultsHTML)
			jobs.GET("/:id/delete", controller.DeleteJobByID)
			jobs.POST("/:id/upload", controller.UploadFilesForJob)
		}

		// all templates API's require auth
		templates := rootRouter.Group("/templates", oauth.Middleware())
		{
			controller := www.NewTemplatesController(flags.TemplatesDirectory, flags.TessdataDirectory)
			templates.GET("/", controller.ListTemplates)
			templates.GET("/new", controller.NewTemplateForm)
			templates.POST("/new", controller.NewTemplatePost)
			templates.GET("/:session", controller.EditTemplateByID)
			templates.POST("/:session", controller.ExportTemplateByID)
			templates.GET("/:session/image", controller.GetTemplateImage)
			templates.POST("/:session/scan", controller.TestTemplateByID)
			templates.POST("/:session/add-area", controller.AddAreaOnTemplate)
			templates.POST("/:session/add-checkpoint", controller.AddCheckpointOnTemplate)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		c.Redirect(307, "/jobs")
	})

	if flags.TLS && len(flags.TLSDomain) > 0 {
		domains := strings.Split(flags.TLSDomain, ",")
		log.Infof("Starting Autocert mode on TLS: %v", domains)
		cacheDir, _ := os.UserCacheDir()
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domains...),
			Cache:      autocert.DirCache(cacheDir),
		}
		log.Fatal(runWithAutocertManager(router, &m))
	} else {
		log.Infof("Starting in plain HTTP on port: %v", flags.ListenPort)
		log.Fatal(router.Run(fmt.Sprintf(":%d", flags.ListenPort)))
	}
}

func runWithAutocertManager(r http.Handler, m *autocert.Manager) error {
	tlsConfig := m.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12

	s := &http.Server{
		Addr:      ":https",
		TLSConfig: tlsConfig,
		Handler:   r,
	}

	l, err := activation.ListenersWithNames()
	if err == nil && len(l) >= 2 {
		log.Info("Running with Unix activation listeners")
		go http.Serve(l["http"][0], m.HTTPHandler(http.HandlerFunc(redirect)))
		return s.ServeTLS(l["https"][0], "", "")
	}

	log.Infof("Starting HTTP Listeners")
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
