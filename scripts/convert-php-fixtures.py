#!/usr/bin/env python3
"""Convert .old-php burnout / en fixtures into YMAIL-shaped JSON packs."""

from __future__ import annotations

import html
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BASE = ROOT / ".old-php" / "fixtures"
OUT_RU = ROOT / "internal" / "adapters" / "fixture" / "data"
OUT_EN = ROOT / "internal" / "adapters" / "fixture" / "data-en"


def strip_html(s: str) -> list[str]:
    if not s:
        return []
    parts = re.split(r"</(?:p|li|h[1-6]|div)>", s, flags=re.I)
    out: list[str] = []
    for p in parts:
        t = re.sub(r"<[^>]+>", " ", p)
        t = html.unescape(t)
        t = re.sub(r"\s+", " ", t).strip()
        if t:
            out.append(t)
    return out


def read_mins(s: str) -> int:
    m = re.search(r"(\d+)", s or "")
    return int(m.group(1)) if m else 5


def glyph_for(pid: str) -> str:
    return {"defense": "◆", "idp": "▣", "ui": "◇", "sdk": "○"}.get(pid, "○")


def write_pack(src: Path, dest: Path, *, lang: str) -> None:
    dest.mkdir(parents=True, exist_ok=True)
    landing = json.loads((src / "landing.json").read_text(encoding="utf-8"))
    site = json.loads((src / "site.json").read_text(encoding="utf-8"))
    about = json.loads((src / "about.json").read_text(encoding="utf-8"))
    projects = json.loads((src / "projects.json").read_text(encoding="utf-8"))
    blog = json.loads((src / "blog.json").read_text(encoding="utf-8"))
    contact = json.loads((src / "contact.json").read_text(encoding="utf-8"))

    brand = {
        "name": site["name"],
        "legalName": site.get("copyright", "© Auth Fly").replace("© ", ""),
        "tagline": landing["hero"]["subtitle"],
        "offer": landing["hero"]["eyebrow"],
        "description": landing["hero"]["description"].strip(),
    }
    (dest / "brand.json").write_text(
        json.dumps(brand, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    advantages = [
        {
            "title": p["label"],
            "description": p["desc"],
            "glyph": glyph_for(p["id"]),
        }
        for p in landing.get("pillars", [])
    ]
    (dest / "advantages.json").write_text(
        json.dumps(advantages, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    accents = ["brand", "highlight", "neutral", "brand", "neutral"]
    section_title = "О модуле" if lang == "ru" else "About this module"
    metric_label = "стек" if lang == "ru" else "stack"
    solutions = []
    for i, pr in enumerate(projects.get("projects", [])):
        solutions.append(
            {
                "id": pr["id"],
                "slug": pr["id"],
                "title": pr["name"],
                "eyebrow": pr.get("category", ""),
                "summary": pr.get("summary", ""),
                "hero": pr.get("tagline", ""),
                "kind": pr["id"],
                "status": pr.get("status", ""),
                "accent": accents[i % len(accents)],
                "outcomes": pr.get("metrics", []),
                "sections": [
                    {"title": section_title, "body": pr.get("description", "").strip()}
                ],
                "metrics": [
                    {"value": t, "label": metric_label, "note": ""}
                    for t in pr.get("tech", [])[:3]
                ],
                "stack": pr.get("tech", []),
            }
        )
    (dest / "solutions.json").write_text(
        json.dumps(solutions, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    posts = []
    for i, p in enumerate(blog.get("posts", [])):
        body = strip_html(p.get("body_html", ""))
        if not body and p.get("excerpt"):
            body = [p["excerpt"]]
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
            }
        )
    (dest / "blog-posts.json").write_text(
        json.dumps(posts, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    about_out = {
        "title": about["hero"]["title"],
        "lead": (
            about["hero"].get("subtitle", "") + " " + about["bio"].get("short", "")
        ).strip(),
        "body": about["bio"].get("long", []),
        "pillars": [
            f'{v["title"]}: {v["desc"]}' for v in about.get("values", [])
        ],
    }
    (dest / "about.json").write_text(
        json.dumps(about_out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    channels = [
        {
            "id": ch["id"],
            "title": ch["label"],
            "value": ch.get("handle", ch.get("url", "")),
            "note": ch.get("desc", ""),
        }
        for ch in contact.get("channels", [])
    ]
    (dest / "contacts.json").write_text(
        json.dumps(channels, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    if lang == "ru":
        legal = [
            {
                "id": "privacy",
                "title": "Политика конфиденциальности",
                "shortTitle": "Конфиденциальность",
                "version": "1.0",
                "effectiveFrom": "2026-08-01",
                "lead": "Сайт не собирает персональные данные форм и не использует счётчики посещений.",
                "sections": [
                    {
                        "id": "data",
                        "title": "Какие данные",
                        "summary": "Регистрации и аналитики на маркетинговом сайте нет.",
                        "points": [
                            "Страницы одинаковы для всех посетителей.",
                            "Тема оформления хранится только в вашем браузере.",
                            "Если появятся формы — политика будет обновлена заранее.",
                        ],
                    }
                ],
            },
            {
                "id": "terms",
                "title": "Условия использования материалов",
                "shortTitle": "Условия",
                "version": "1.0",
                "effectiveFrom": "2026-08-01",
                "lead": "Материалы справочные; MIT-код Auth Fly — в репозиториях на GitHub.",
                "sections": [
                    {
                        "id": "scope",
                        "title": "Область",
                        "summary": "Описания модулей не являются договором на поддержку.",
                        "points": [
                            "Сверяйте версии с документацией репозиториев.",
                            "Ответственность за деплой несёт оператор сервера.",
                        ],
                    }
                ],
            },
        ]
    else:
        legal = [
            {
                "id": "privacy",
                "title": "Privacy policy",
                "shortTitle": "Privacy",
                "version": "1.0",
                "effectiveFrom": "2026-08-01",
                "lead": "This marketing site does not collect form PII and does not run analytics counters.",
                "sections": [
                    {
                        "id": "data",
                        "title": "What we process",
                        "summary": "No registration or analytics on the marketing site.",
                        "points": [
                            "Pages are the same for every visitor.",
                            "Theme preference stays in your browser only.",
                            "If forms appear later, this policy will be updated first.",
                        ],
                    }
                ],
            },
            {
                "id": "terms",
                "title": "Terms of use",
                "shortTitle": "Terms",
                "version": "1.0",
                "effectiveFrom": "2026-08-01",
                "lead": "Materials are informational; Auth Fly MIT code lives in the GitHub repositories.",
                "sections": [
                    {
                        "id": "scope",
                        "title": "Scope",
                        "summary": "Module descriptions are not a support contract.",
                        "points": [
                            "Check repository docs for the version you deploy.",
                            "The server operator is responsible for deployment.",
                        ],
                    }
                ],
            },
        ]
    (dest / "legal.json").write_text(
        json.dumps(legal, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )

    print(f"{lang}: posts={len(posts)} modules={len(solutions)} -> {dest}")


def main() -> None:
    write_pack(BASE / "voice" / "burnout", OUT_RU, lang="ru")
    write_pack(BASE / "en", OUT_EN, lang="en")


if __name__ == "__main__":
    main()
