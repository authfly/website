package site

import (
	"net/http"

	"github.com/fastygo/framework/pkg/app"
	"github.com/authfly/website/internal/ports"
)

// Deps are the ports the informational site feature needs.
type Deps struct {
	Content ports.Content
}

// Feature is the SSR AuthFly marketing site (ID "site").
type Feature struct {
	deps Deps
	nav  []app.NavItem
}

// New constructs the site feature.
func New(deps Deps) *Feature {
	return &Feature{deps: deps}
}

// ID implements app.Feature.
func (f *Feature) ID() string { return "site" }

// NavItems implements app.Feature.
func (f *Feature) NavItems() []app.NavItem {
	return []app.NavItem{
		{Label: "Модули", Path: "/projects", Icon: "layers", Order: 10},
		{Label: "Журнал", Path: "/blog", Icon: "book", Order: 20},
		{Label: "О проекте", Path: "/about", Icon: "info", Order: 30},
		{Label: "Контакты", Path: "/contacts", Icon: "phone", Order: 40},
	}
}

// SetNavItems implements app.NavProvider.
func (f *Feature) SetNavItems(items []app.NavItem) {
	f.nav = items
}

// Routes registers every informational path on mux.
func (f *Feature) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", f.handleHome)
	mux.HandleFunc("GET /projects", f.handleSolutions)
	mux.HandleFunc("GET /projects/{slug}", f.handleSolution)
	mux.HandleFunc("GET /blog", f.handleBlog)
	mux.HandleFunc("GET /blog/{slug}", f.handleBlogPost)
	mux.HandleFunc("GET /about", f.handleAbout)
	mux.HandleFunc("GET /contacts", f.handleContacts)
	mux.HandleFunc("GET /contact", f.handleContacts) // PHP alias
	mux.HandleFunc("GET /legal/{doc}", f.handleLegal)
}
