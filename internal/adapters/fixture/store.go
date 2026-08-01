package fixture

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/authfly/website/internal/ports"
)

// Store is the fixture-backed content adapter (JSON now → DB later).
type Store struct {
	mu sync.RWMutex

	brand      ports.Brand
	advantages []ports.Advantage
	solutions  []ports.Solution
	bySolution map[string]ports.Solution
	posts      []ports.BlogPost
	byPost     map[string]ports.BlogPost
	about      ports.AboutPage
	channels   []ports.ContactChannel
	docs       []ports.LegalDocument
	byDoc      map[string]ports.LegalDocument
}

// Load reads JSON fixtures from dir (default: internal/adapters/fixture/data).
func Load(dir string) (*Store, error) {
	if dir == "" {
		dir = filepath.Join("internal", "adapters", "fixture", "data")
	}
	s := &Store{
		bySolution: map[string]ports.Solution{},
		byPost:     map[string]ports.BlogPost{},
		byDoc:      map[string]ports.LegalDocument{},
	}
	must := func(name string, dest any) error {
		raw, err := os.ReadFile(filepath.Join(dir, name))
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

func (s *Store) Brand(context.Context) (ports.Brand, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.brand, nil
}

func (s *Store) Advantages(context.Context) ([]ports.Advantage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Advantage, len(s.advantages))
	copy(out, s.advantages)
	return out, nil
}

func (s *Store) Solutions(context.Context) ([]ports.Solution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Solution, len(s.solutions))
	copy(out, s.solutions)
	return out, nil
}

func (s *Store) SolutionBySlug(_ context.Context, slug string) (*ports.Solution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.bySolution[slug]
	if !ok {
		return nil, nil
	}
	cp := item
	return &cp, nil
}

func (s *Store) BlogPosts(context.Context) ([]ports.BlogPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.BlogPost, len(s.posts))
	copy(out, s.posts)
	return out, nil
}

func (s *Store) BlogPostBySlug(_ context.Context, slug string) (*ports.BlogPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.byPost[slug]
	if !ok {
		return nil, nil
	}
	cp := item
	return &cp, nil
}

func (s *Store) About(context.Context) (ports.AboutPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.about, nil
}

func (s *Store) ContactChannels(context.Context) ([]ports.ContactChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.ContactChannel, len(s.channels))
	copy(out, s.channels)
	return out, nil
}

func (s *Store) LegalDocument(_ context.Context, id string) (*ports.LegalDocument, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.byDoc[id]
	if !ok {
		return nil, nil
	}
	cp := item
	return &cp, nil
}

var _ ports.Content = (*Store)(nil)
