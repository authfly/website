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
	layout, err := f.layout(ctx, r, brand.Name+" — "+brand.Tagline)
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	advantages, _ := f.deps.Content.Advantages(ctx)
	solutions, _ := f.deps.Content.Solutions(ctx)
	posts, _ := f.deps.Content.BlogPosts(ctx)

	cards := f.solutionCards(solutions)
	var featured *viewmodels.SolutionCardModel
	if len(cards) > 0 {
		featured = &cards[0]
		cards = cards[1:]
	}

	f.render(w, r, views.HomePage(viewmodels.HomeModel{
		Layout:     layout,
		Advantages: advantages,
		Featured:   featured,
		Solutions:  take(cards, 4),
		Posts:      take(f.blogCards(featuredPosts(posts)), 2),
	}))
}

func (f *Feature) handleSolutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	pages := layoutPages(layoutLang()).Pages
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.Projects, brand.Name))
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
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name))
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(sol.Title, brand.Name))
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
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.Blog, brand.Name))
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
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name))
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(post.Title, brand.Name))
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
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.About, brand.Name))
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	page, _ := f.deps.Content.About(ctx)
	f.render(w, r, views.AboutPage(viewmodels.AboutModel{
		Layout: layout,
		Page:   page,
		Brand:  brand,
	}))
}

func (f *Feature) handleContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	brand, _ := f.deps.Content.Brand(ctx)
	pages := layoutPages(layoutLang()).Pages
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(pages.Contacts, brand.Name))
	if err != nil {
		http.Error(w, "layout error", http.StatusInternalServerError)
		return
	}
	channels, _ := f.deps.Content.ContactChannels(ctx)
	f.render(w, r, views.ContactsPage(viewmodels.ContactsModel{
		Layout:   layout,
		Channels: channels,
	}))
}

func (f *Feature) handleLegal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := r.PathValue("doc")
	brand, _ := f.deps.Content.Brand(ctx)
	doc, err := f.deps.Content.LegalDocument(ctx, docID)
	if err != nil || doc == nil {
		nf := layoutPages(layoutLang()).Pages.NotFound
		layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name))
		w.WriteHeader(http.StatusNotFound)
		f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
		return
	}
	layout, err := f.layout(ctx, r, viewmodels.WithBrand(doc.ShortTitle, brand.Name))
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
	layout, _ := f.layout(ctx, r, viewmodels.WithBrand(nf, brand.Name))
	w.WriteHeader(http.StatusNotFound)
	f.render(w, r, views.NotFoundPage(viewmodels.NotFoundModel{Layout: layout, Path: r.URL.Path}))
}
