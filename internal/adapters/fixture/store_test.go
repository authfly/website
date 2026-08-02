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
	home, err := store.HomeMeta(ctx)
	if err != nil || len(home.Path) == 0 || home.Portrait.Src == "" {
		t.Fatalf("HomeMeta: %#v err=%v", home, err)
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
	contacts, err := store.Contacts(ctx)
	if err != nil || len(contacts.Channels) == 0 || contacts.Channels[0].Href == "" {
		t.Fatalf("Contacts: %#v err=%v", contacts, err)
	}
	doc, err := store.LegalDocument(ctx, "privacy")
	if err != nil || doc == nil {
		t.Fatalf("LegalDocument: %#v err=%v", doc, err)
	}
}

func TestLoadEN(t *testing.T) {
	t.Parallel()
	store, err := Load(filepath.Join("data-en"))
	if err != nil {
		t.Fatalf("Load EN: %v", err)
	}
	ctx := context.Background()
	brand, err := store.Brand(ctx)
	if err != nil || brand.Name == "" {
		t.Fatalf("Brand EN: %#v err=%v", brand, err)
	}
	home, err := store.HomeMeta(ctx)
	if err != nil || len(home.Evidence) == 0 {
		t.Fatalf("HomeMeta EN: %#v err=%v", home, err)
	}
	contacts, err := store.Contacts(ctx)
	if err != nil || contacts.Title == "" {
		t.Fatalf("Contacts EN: %#v err=%v", contacts, err)
	}
}

func TestLoadLocaleEmbed(t *testing.T) {
	t.Parallel()
	ru, err := LoadLocale("ru")
	if err != nil {
		t.Fatalf("embed ru: %v", err)
	}
	en, err := LoadLocale("en")
	if err != nil {
		t.Fatalf("embed en: %v", err)
	}
	if _, err := ru.HomeMeta(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := en.Contacts(context.Background()); err != nil {
		t.Fatal(err)
	}
}
