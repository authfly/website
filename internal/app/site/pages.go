package site

import (
	"net/http"

	"github.com/authfly/website/internal/viewmodels"
	"github.com/authfly/website/internal/views"
)

func (f *Feature) handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, err := f.deps.Content.Brand(ctx)
	if err != nil {
		http.Error(w, "content error", http.StatusInternalServerError)
		return
	}
	meta, _ := f.deps.Content.HomeMeta(ctx)
	title := brand.Name + " — " + brand.Tagline
	desc := brand.Description
	if meta.SEO.Title != "" {
		title = meta.SEO.Title
	}
	if meta.SEO.Description != "" {
		desc = meta.SEO.Description
	}
	layout, err := f.layout(ctx, r, title, desc)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	advantages, _ := f.deps.Content.Advantages(ctx)
	solutions, _ := f.deps.Content.Solutions(ctx)
	posts, _ := f.deps.Content.BlogPosts(ctx)

	f.render(w, r, views.HomePage(viewmodels.HomeModel{
		Layout:     layout,
		Meta:       meta,
		Advantages: advantages,
		Solutions:  take(f.solutionCards(solutions), 4),
		Posts:      take(f.blogCards(featuredPosts(posts)), 2),
	}))
}

func (f *Feature) handleSolutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	pages := layoutPages(layoutLang()).Pages
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.Projects, brand.Name), brand.Description)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	list, _ := f.deps.Content.Solutions(ctx)
	f.render(w, r, views.SolutionsPage(viewmodels.SolutionsModel{
		Layout:    layout,
		Solutions: f.solutionCards(list),
	}))
}

func (f *Feature) handleSolution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("slug")
	brand, _ := f.deps.Content.Brand(ctx)
	sol, err := f.deps.Content.SolutionBySlug(ctx, slug)
	if err != nil || sol == nil {
		nf := layoutPages(layoutLang()).Pages.NotFound
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name), brand.Description)
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(sol.Title, brand.Name), sol.Summary)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	f.render(w, r, views.SolutionPage(viewmodels.SolutionModel{
		Layout:   layout,
		Solution: *sol,
	}))
}

func (f *Feature) handleBlog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	pages := layoutPages(layoutLang()).Pages
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.Blog, brand.Name), brand.Description)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	posts, _ := f.deps.Content.BlogPosts(ctx)
	f.render(w, r, views.BlogPage(viewmodels.BlogModel{
		Layout: layout,
		Posts:  f.blogCards(posts),
	}))
}

func (f *Feature) handleBlogPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("slug")
	brand, _ := f.deps.Content.Brand(ctx)
	post, err := f.deps.Content.BlogPostBySlug(ctx, slug)
	if err != nil || post == nil {
		nf := layoutPages(layoutLang()).Pages.NotFound
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name), brand.Description)
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(post.Title, brand.Name), post.Excerpt)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	f.render(w, r, views.BlogPostPage(viewmodels.BlogPostModel{
		Layout: layout,
		Post:   *post,
	}))
}

func (f *Feature) handleAbout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	pages := layoutPages(layoutLang()).Pages
	page, _ := f.deps.Content.About(ctx)
	title := viewmodels.WithBrand(pages.About, brand.Name)
	desc := page.Lead
	if page.SEO.Title != "" {
		title = page.SEO.Title
	}
	if page.SEO.Description != "" {
		desc = page.SEO.Description
	}
	layout, err := f.layout(ctx, r, title, desc)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	f.render(w, r, views.AboutPage(viewmodels.AboutModel{
		Layout: layout,
		Page:   page,
		Brand:  brand,
	}))
}

func (f *Feature) handleContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, _ := f.deps.Content.Contacts(ctx)
	title := page.Title
	if title == "" {
		title = layoutPages(layoutLang()).Pages.Contacts
	}
	desc := page.Lead
	if page.SEO.Description != "" {
		desc = page.SEO.Description
	}
	if page.SEO.Title != "" {
		title = page.SEO.Title
	}
	layout, err := f.layout(ctx, r, title, desc)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	f.render(w, r, views.ContactsPage(viewmodels.ContactsModel{
		Layout: layout,
		Page:   page,
	}))
}

func (f *Feature) handleLegal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := r.PathValue("doc")
	brand, _ := f.deps.Content.Brand(ctx)
	doc, err := f.deps.Content.LegalDocument(ctx, docID)
	if err != nil || doc == nil {
		nf := layoutPages(layoutLang()).Pages.NotFound
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name), brand.Description)
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(doc.ShortTitle, brand.Name), doc.Lead)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	f.render(w, r, views.LegalPage(viewmodels.LegalModel{Layout: layout, Doc: *doc}))
}

func (f *Feature) handleNotFound(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	nf := layoutPages(layoutLang()).Pages.NotFound
	layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name), brand.Description)
	w.WriteHeader(http.StatusNotFound)
	f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
}
