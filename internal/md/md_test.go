package md

import "testing"

func TestParse(t *testing.T) {
	d := Parse("---\ntitle: T\ndescription: D\n---\n# Head `x`\n\nPara with **b** and [l](/en/a) and `c`.\nmore\n\n- one\n  cont\n- two\n\n1. a\n2. b\n\n```sh\necho | hi\n```\n\n| A | B |\n|---|---|\n| `a|b` | x\\|y |\n\n> **Note:** n\n\n---\n## Head `x`\n")
	if d.Meta["title"] != "T" || d.Meta["description"] != "D" {
		t.Fatalf("meta %v", d.Meta)
	}
	kinds := []string{Heading, Paragraph, Bullets, Numbers, Code, Table, Note, Rule, Heading}
	if len(d.Blocks) != len(kinds) {
		t.Fatalf("blocks %d: %+v", len(d.Blocks), d.Blocks)
	}
	for i, k := range kinds {
		if d.Blocks[i].Kind != k {
			t.Fatalf("block %d kind %s want %s", i, d.Blocks[i].Kind, k)
		}
	}
	if d.Blocks[0].ID != "head-x" || d.Blocks[8].ID != "head-x-1" {
		t.Fatalf("ids %s %s", d.Blocks[0].ID, d.Blocks[8].ID)
	}
	if p := d.Blocks[1].Text; len(p) != 7 || p[1].Kind != Strong || p[3].Kind != Link || p[3].Href != "/en/a" || p[5].Kind != Mono {
		t.Fatalf("inlines %+v", p)
	}
	if it := d.Blocks[2].Items; len(it) != 2 || Plain(it[0]) != "one cont" {
		t.Fatalf("list %+v", it)
	}
	if d.Blocks[4].Code != "echo | hi" || d.Blocks[4].Lang != "sh" {
		t.Fatalf("code %+v", d.Blocks[4])
	}
	if r := d.Blocks[5].Rows; len(r) != 1 || len(r[0]) != 2 || r[0][0][0].Text != "a|b" || Plain(r[0][1]) != "x|y" {
		t.Fatalf("table %+v", d.Blocks[5])
	}
	// Cyrillic letters survive (escaped here: no Cyrillic literals in code)
	if got := Slug("\u0423\u0441 Authfly + Stalwart"); got != "\u0443\u0441-authfly-stalwart" {
		t.Fatal(got)
	}
}
