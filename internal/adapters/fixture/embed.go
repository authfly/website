package fixture

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/authfly/website/internal/ports"
)

//go:embed data/*.json
var dataRU embed.FS

//go:embed data-en/*.json
var dataEN embed.FS

// LoadLocale loads the embedded RU or EN fixture pack (Vercel-safe).
func LoadLocale(locale string) (*Store, error) {
	switch locale {
	case "en":
		return loadFS(dataEN, "data-en")
	default:
		return loadFS(dataRU, "data")
	}
}

func loadFS(fsys embed.FS, root string) (*Store, error) {
	s := &Store{
		bySolution: map[string]ports.Solution{},
		byPost:     map[string]ports.BlogPost{},
		byDoc:      map[string]ports.LegalDocument{},
	}
	must := func(name string, dest any) error {
		raw, err := fs.ReadFile(fsys, root+"/"+name)
		if err != nil {
			return fmt.Errorf("fixture %s: %w", name, err)
		}
		if err := json.Unmarshal(raw, dest); err != nil {
			return fmt.Errorf("fixture %s decode: %w", name, err)
		}
		return nil
	}
	if err := must("brand.json", &s.brand); err != nil {
		return nil, err
	}
	if err := must("advantages.json", &s.advantages); err != nil {
		return nil, err
	}
	if err := must("solutions.json", &s.solutions); err != nil {
		return nil, err
	}
	if err := must("blog-posts.json", &s.posts); err != nil {
		return nil, err
	}
	if err := must("about.json", &s.about); err != nil {
		return nil, err
	}
	if err := must("contacts.json", &s.channels); err != nil {
		return nil, err
	}
	if err := must("legal.json", &s.docs); err != nil {
		return nil, err
	}
	for _, item := range s.solutions {
		s.bySolution[item.Slug] = item
	}
	for _, item := range s.posts {
		s.byPost[item.Slug] = item
	}
	for _, item := range s.docs {
		s.byDoc[item.ID] = item
	}
	return s, nil
}
