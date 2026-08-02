#!/usr/bin/env python3
"""Regenerate blog-posts + about fixtures from PHP archive (git 8dd879d sources in .manual/)."""

from __future__ import annotations

import html
import json
import re
from html.parser import HTMLParser
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


class BodyParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.blocks: list[dict] = []
        self._tag: str | None = None
        self._buf: list[str] = []
        self._li: list[str] = []
        self._in_li = False
        self._li_buf: list[str] = []

    def handle_starttag(self, tag: str, attrs) -> None:
        tag = tag.lower()
        if tag in ("p", "h2", "h3"):
            self._tag = tag
            self._buf = []
        elif tag == "ul":
            self._tag = "ul"
            self._li = []
        elif tag == "li" and self._tag == "ul":
            self._in_li = True
            self._li_buf = []
        elif tag == "br":
            if self._in_li:
                self._li_buf.append(" ")
            elif self._tag:
                self._buf.append(" ")

    def handle_endtag(self, tag: str) -> None:
        tag = tag.lower()
        if tag in ("p", "h2", "h3") and self._tag == tag:
            text = html.unescape(re.sub(r"\s+", " ", "".join(self._buf))).strip()
            if text:
                kind = "p" if tag == "p" else tag
                self.blocks.append({"kind": kind, "text": text})
            self._tag = None
            self._buf = []
        elif tag == "li" and self._in_li:
            text = html.unescape(re.sub(r"\s+", " ", "".join(self._li_buf))).strip()
            if text:
                self._li.append(text)
            self._in_li = False
            self._li_buf = []
        elif tag == "ul" and self._tag == "ul":
            if self._li:
                self.blocks.append({"kind": "ul", "items": self._li})
            self._tag = None
            self._li = []

    def handle_data(self, data: str) -> None:
        if self._in_li:
            self._li_buf.append(data)
        elif self._tag in ("p", "h2", "h3"):
            self._buf.append(data)


def read_mins(s: str) -> int:
    m = re.search(r"(\d+)", s or "")
    return int(m.group(1)) if m else 5


def clean_url(url: str) -> str:
    url = (url or "").split("?")[0].rstrip("/")
    if url.startswith("/blog/"):
        return url
    return url


def convert_blog(src: Path, dest: Path) -> None:
    blog = json.loads(src.read_text(encoding="utf-8"))
    posts = []
    for i, p in enumerate(blog.get("posts", [])):
        parser = BodyParser()
        parser.feed(p.get("body_html") or "")
        blocks = parser.blocks
        if not blocks and p.get("excerpt"):
            blocks = [{"kind": "p", "text": p["excerpt"]}]
        # Keep flat body for any legacy readers / search snippets
        body = []
        for b in blocks:
            if b["kind"] == "ul":
                body.extend(b.get("items") or [])
            elif b.get("text"):
                body.append(b["text"])
        posts.append(
            {
                "slug": p["slug"],
                "title": p["title"],
                "excerpt": p.get("excerpt", ""),
                "date": p.get("date", ""),
                "readMinutes": read_mins(p.get("reading_time", "5")),
                "tags": [p.get("tag", p.get("category", ""))],
                "featured": i < 2,
                "body": body,
                "blocks": blocks,
            }
        )
    dest.write_text(json.dumps(posts, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"blog {len(posts)} -> {dest}")


def convert_about(src: Path, dest: Path) -> None:
    about = json.loads(src.read_text(encoding="utf-8"))
    books = []
    for b in about.get("books", []):
        books.append(
            {
                "title": b["title"],
                "publisher": b.get("publisher", ""),
                "year": b.get("year", ""),
                "pages": b.get("pages", 0),
                "series": b.get("series", ""),
                "desc": b.get("desc", ""),
                "href": clean_url(b.get("url", "")),
                "cta": b.get("cta", ""),
            }
        )
    out = {
        "seo": about.get("seo", {}),
        "title": about["hero"]["title"],
        "lead": (about["hero"].get("subtitle", "") + " " + about["bio"].get("short", "")).strip(),
        "body": about["bio"].get("long", []),
        "pillars": [f'{v["title"]}: {v["desc"]}' for v in about.get("values", [])],
        "pillars_title": about.get("sections", {}).get("principles_title", ""),
        "books_eyebrow": about.get("sections", {}).get("books", {}).get("eyebrow", ""),
        "books_title": about.get("sections", {}).get("books", {}).get("title", ""),
        "books_intro": about.get("sections", {}).get("books", {}).get("intro", ""),
        "timeline_title": about.get("sections", {}).get("timeline_title", ""),
        "portrait": {
            "src": "/assets/img/senior.png",
            "aria": about.get("hero_portrait", {}).get("aria_label", ""),
        },
        "meta": about.get("sidebar_meta", []),
        "books": books,
        "timeline": [
            {
                "period": t["period"],
                "label": t["label"],
                "detail": t["detail"],
            }
            for t in about.get("timeline", [])
        ],
    }
    dest.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"about books={len(books)} timeline={len(out['timeline'])} -> {dest}")


def main() -> None:
    convert_blog(
        ROOT / ".manual/burnout-src/blog.json",
        ROOT / "internal/adapters/fixture/data/blog-posts.json",
    )
    convert_blog(
        ROOT / ".manual/en-src/blog.json",
        ROOT / "internal/adapters/fixture/data-en/blog-posts.json",
    )
    convert_about(
        ROOT / ".manual/burnout-src/about.json",
        ROOT / "internal/adapters/fixture/data/about.json",
    )
    convert_about(
        ROOT / ".manual/en-src/about.json",
        ROOT / "internal/adapters/fixture/data-en/about.json",
    )


if __name__ == "__main__":
    main()
