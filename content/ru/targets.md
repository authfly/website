---
title: Платформы
description: Девять рантаймов виджетов Authfly, как поставить каждый из реестра UI8Kit и какой серверный адаптер к нему подходит.
---
# Платформы: одна система, 9 платформ

Каждая платформа рендерит для шести виджетов одинаковые DOM, классы и ARIA. CI authkit рендерит 20 эталонных состояний на каждом рантайме и сравнивает HTML с React (паритет), а также ставит каждый элемент реестра в пустой проект для каждого рантайма (смоук реестра).

| Платформа | Каталог установки | `app.runtime` | Типичный `app.adapter` | Пакет в authkit |
|---|---|---|---|---|
| React (эталон) | `src` | `react` | `node` / `bun` | `packages/ui/react` |
| Vue 3 | `src` | `vue` | `node` / `bun` | `packages/ui/vue` |
| Svelte 5 | `src` | `svelte` | `node` / `bun` | `packages/ui/svelte` |
| Solid | `src` | `solid` | `node` / `bun` | `packages/ui/solid` |
| Web Components | `src` | не принимается | любой | `packages/ui/webcomponent` |
| templ (Go) | `.` | `templ` | `go` | `packages/ui/templ` |
| Twig (PHP) | `.` | `twig` | `php` | `packages/ui/twig`, зеркало authfly/twig |
| Latte (PHP) | `.` | `latte` | `php` | `packages/ui/latte`, зеркало authfly/latte |
| Jinja2 (Python) | `.` | `jinja2` | `python` | `packages/ui/jinja2` |

## Установка

```sh
npx ui8kit@3.2.1 init -y --framework <target> --dir <install dir> --registry-version 3.3.0
npx ui8kit@3.2.1 add authfly-sign-in authfly-sign-up authfly-account \
  --registry-url https://registry.buildy.tw/authfly/<target>
```

Для webcomponent и templ добавляйте `authfly-widgets` (все шесть виджетов одним файлом) вместо отдельных виджетов.

> **Не проверено:** для этой документации вручную выполнена только установка React (ноябрь 2026). Остальные платформы покрыты смоуком реестра в CI authkit, который ставит из локальной копии реестра с примитивами с unpkg; именно эти флаги `init` для них не запускались.

## Заметки по платформам

- **Twig, Latte:** с каталогом установки `.` CLI UI8Kit (3.4.0) кладёт `authfly-lib` (`src/*.php`) в корень проекта; верните файлы в `src/` или настройте namespace (`Authfly\Twig\`, `Authfly\Latte\`). Autoload Composer CLI не пишет.
- **templ:** Go-файлы Authfly импортируют `github.com/authfly/authkit/packages/ui/templ/...`. CLI переписывает пути импорта только для примитивов UI8Kit, поэтому либо используйте модуль authkit templ напрямую (`go get github.com/authfly/authkit/packages/ui/templ`), либо перепишите импорты. Сгенерированные `*_templ.go` поставляются готовыми (templ v0.3.1070).
- **Jinja2:** `authfly_jinja` ставится файлами; добавьте его в свой `requirements.txt` или `pyproject.toml`.
- **Web Components:** `ui8kit init --framework webcomponent` требует `UI8KIT_EXPERIMENTAL=webcomponent`; в реестре Authfly нет сборки webcomponent, примитивы берутся из `@ui8kit/registry@3.3.0` на unpkg.
- Зависимости фреймворка (`react`, `vue`, `twig/twig`, …) элементы реестра не объявляют; их приносит ваше приложение.

## Серверная часть

Виджеты — это формы; с Authfly общается ваш сервер. Используйте SDK вашего адаптера ([SDK](/ru/reference/sdks)) или любой OIDC-клиент по стандарту. [Гайд React + Docker](/ru/guides/react-docker) показывает полный шаблон для семейства JavaScript.
