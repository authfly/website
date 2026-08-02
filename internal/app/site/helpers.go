package site

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/fastygo/framework/pkg/web"
	"github.com/authfly/website/internal/fixtures"
	"github.com/authfly/website/internal/ports"
	"github.com/authfly/website/internal/viewmodels"
)

const defaultOgImage = "/assets/img/authfly.png"

func layoutLang() string {
	return fixtures.ActiveLocale()
}

func layoutPages(lang string) fixtures.Locale {
	loc, err := fixtures.LoadLocale(lang)
	if err != nil {
		return fixtures.MustLocale("ru")
	}
	return loc
}

func (f *Feature) layout(ctx context.Context, r *http.Request, title, description string) (viewmodels.LayoutModel, error) {
	brand, err := f.deps.Content.Brand(ctx)
	if err != nil {
		return viewmodels.LayoutModel{}, err
	}
	lang := layoutLang()
	copy, err := fixtures.LoadLocale(lang)
	if err != nil {
		return viewmodels.LayoutModel{}, err
	}
	if description == "" {
		description = brand.Description
	}
	return viewmodels.LayoutModel{
		Title:       title,
		Description: description,
		OgImage:     defaultOgImage,
		Canonical:   r.URL.Path,
		Brand:       brand,
		Offer:       brand.Offer,
		Path:        r.URL.Path,
		Copy:        copy,
		Lang:        lang,
	}, nil
}

func (f *Feature) solutionCards(list []ports.Solution) []viewmodels.SolutionCardModel {
	out := make([]viewmodels.SolutionCardModel, 0, len(list))
	for _, p := range list {
		out = append(out, viewmodels.SolutionCardModel{
			Slug:    p.Slug,
			Title:   p.Title,
			Eyebrow: p.Eyebrow,
			Summary: p.Summary,
			Status:  p.Status,
			Accent:  p.Accent,
			Href:    "/projects/" + p.Slug,
		})
	}
	return out
}

func take[T any](list []T, n int) []T {
	if len(list) > n {
		return list[:n]
	}
	return list
}

func featuredPosts(list []ports.BlogPost) []ports.BlogPost {
	out := make([]ports.BlogPost, 0, len(list))
	for _, p := range list {
		if p.Featured {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return list
	}
	return out
}

func (f *Feature) blogCards(list []ports.BlogPost) []viewmodels.BlogCardModel {
	lang := layoutLang()
	out := make([]viewmodels.BlogCardModel, 0, len(list))
	for _, p := range list {
		out = append(out, viewmodels.BlogCardModel{
			Slug:        p.Slug,
			Title:       p.Title,
			Excerpt:     p.Excerpt,
			Date:        p.Date,
			ReadMinutes: p.ReadMinutes,
			Tags:        p.Tags,
			Href:        "/blog/" + p.Slug,
			Featured:    p.Featured,
			Lang:        lang,
		})
	}
	return out
}

func (f *Feature) render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if err := web.Render(r.Context(), w, component); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}
