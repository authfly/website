package fixture

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	dir := filepath.Join("data")
	store, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	ctx := context.Background()
	brand, err := store.Brand(ctx)
	if err != nil || brand.Name == "" {
		t.Fatalf("Brand: %#v err=%v", brand, err)
	}
	solutions, err := store.Solutions(ctx)
	if err != nil || len(solutions) != 5 {
		t.Fatalf("Solutions: len=%d err=%v", len(solutions), err)
	}
	sol, err := store.SolutionBySlug(ctx, "idp")
	if err != nil || sol == nil {
		t.Fatalf("SolutionBySlug: %#v err=%v", sol, err)
	}
	posts, err := store.BlogPosts(ctx)
	if err != nil || len(posts) < 1 {
		t.Fatalf("BlogPosts: len=%d err=%v", len(posts), err)
	}
	doc, err := store.LegalDocument(ctx, "privacy")
	if err != nil || doc == nil {
		t.Fatalf("LegalDocument: %#v err=%v", doc, err)
	}
}

func TestLoadLocaleEN(t *testing.T) {
	t.Parallel()
	store, err := LoadLocale("en")
	if err != nil {
		t.Fatalf("LoadLocale en: %v", err)
	}
	brand, err := store.Brand(context.Background())
	if err != nil || brand.Name == "" {
		t.Fatalf("Brand EN embed: %#v err=%v", brand, err)
	}
}

func TestLoadEN(t *testing.T) {
	t.Parallel()
	store, err := Load(filepath.Join("data-en"))
	if err != nil {
		t.Fatalf("Load EN: %v", err)
	}
	brand, err := store.Brand(context.Background())
	if err != nil || brand.Name == "" {
		t.Fatalf("Brand EN: %#v err=%v", brand, err)
	}
}
