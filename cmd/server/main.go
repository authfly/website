// Command server serves the Authfly docs site (Vercel Go runtime or any host): PORT, SITE_ORIGIN.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/authfly/website/assets"
	"github.com/authfly/website/content"
	"github.com/authfly/website/internal/site"
	"github.com/authfly/website/internal/views"
)

func main() {
	c, err := site.Load(content.FS)
	if err != nil {
		log.Fatal(err)
	}
	if errs := c.Check(); len(errs) > 0 {
		log.Fatalf("content check: %v", errs)
	}
	origin := os.Getenv("SITE_ORIGIN")
	if origin == "" {
		origin = "https://authfly.vercel.app"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3011"
	}
	if sha := os.Getenv("VERCEL_GIT_COMMIT_SHA"); len(sha) >= 7 {
		views.Version = sha[:7]
	}
	s := &site.Server{C: c, Assets: assets.FS, Origin: origin}
	srv := &http.Server{Addr: ":" + port, Handler: s.Handler(), ReadHeaderTimeout: 10 * time.Second}
	log.Printf("authfly website on :%s", port)
	log.Fatal(srv.ListenAndServe())
}
