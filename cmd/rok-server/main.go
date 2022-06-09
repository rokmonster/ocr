package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	config "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config/serverconfig"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/webcontrollers"
	"github.com/xor22h/rok-monster-ocr-golang/web"
	"golang.org/x/crypto/acme/autocert"

	bolt "go.etcd.io/bbolt"
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

	db, err := bolt.Open("db.bolt", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New(fiber.Config{
		TrustedProxies:        []string{},
		Prefork:               false,
		EnablePrintRoutes:     true,
		Views:                 web.CreateTemplateEngine(web.StaticFS, "template"),
		DisableStartupMessage: true,
		BodyLimit:             64 << 20, // 64MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Redirect("/jobs", 307)
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// public group, not auth needed for this.
	public := app.Group("")
	{
		webcontrollers.NewJobsController(public.Group("/jobs"), db).Setup()
		webcontrollers.NewTemplatesController(public.Group("/templates"), flags.TemplatesDirectory, flags.TessdataDirectory).Setup()
	}

	app.Use(filesystem.New(filesystem.Config{
		Root: web.EmbeddedFS(web.StaticFS, "static"),
	}))

	if flags.TLS && len(flags.TLSDomain) > 0 {
		log.Infof("Starting Autocert mode on TLS: https://%v", flags.TLSDomain)
		cacheDir, _ := os.UserCacheDir()
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(flags.TLSDomain),
			Cache:      autocert.DirCache(cacheDir),
		}
		log.Fatal(runWithAutocertManager(app, &m))
	} else {
		log.Infof("Starting in plain HTTP on port: %v", flags.ListenPort)
		log.Fatal(app.Listen(fmt.Sprintf(":%d", flags.ListenPort)))
	}
}

func runWithAutocertManager(r *fiber.App, m *autocert.Manager) error {
	config := m.TLSConfig()
	config.MinVersion = tls.VersionTLS12

	l, err := activation.ListenersWithNames()
	if err == nil && len(l) >= 2 {
		log.Info("Running with Unix activation listeners")
		go http.Serve(l["http"][0], m.HTTPHandler(http.HandlerFunc(redirect)))

		ln := tls.NewListener(l["https"][0], config)
		return r.Listener(ln)
	}

	log.Infof("Starting HTTP Listeners")
	go http.ListenAndServe(":http", m.HTTPHandler(http.HandlerFunc(redirect)))
	tlsListener, _ := tls.Listen("tcp", ":https", config)
	return r.Listener(tlsListener)
}

func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
