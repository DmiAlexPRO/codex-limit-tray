"""Render the GitHub Pages site at build time. GitHub serves the committed HTML."""

import argparse
import json
from html import escape
from pathlib import Path
from string import Template

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
SOURCE = ROOT / "scripts" / "pages"
SITE = "https://dmialexpro.github.io/codex-limit-tray/"
REPOSITORY = "https://github.com/DmiAlexPRO/codex-limit-tray"
DOWNLOAD = f"{REPOSITORY}/releases/latest/download/CodexLimitTray-windows-x64-Setup.exe"
LOCALES = {
    "en": ("English", "en"),
    "es": ("Español", "es"),
    "zh": ("简体中文", "zh-Hans"),
    "ru": ("Русский", "ru"),
    "it": ("Italiano", "it"),
    "de": ("Deutsch", "de"),
    "uk": ("Українська", "uk"),
    "kk": ("Қазақша", "kk"),
    "pl": ("Polski", "pl"),
    "cs": ("Čeština", "cs"),
    "ko": ("한국어", "ko"),
}


def url_path(code):
    return "" if code == "en" else f"{code}/"


def render():
    translations = json.loads((SOURCE / "locales.json").read_text(encoding="utf-8"))
    template = Template((SOURCE / "index.html.tmpl").read_text(encoding="utf-8"))
    keys = translations["en"].keys()
    if translations.keys() != LOCALES.keys():
        raise ValueError("Website locales do not match the list of supported languages")
    for code, messages in translations.items():
        if keys != messages.keys() or not all(isinstance(v, str) and v.strip() for v in messages.values()):
            raise ValueError(f"Missing, invalid or extra translation in {code}")

    alternates = "\n".join(
        f'  <link rel="alternate" hreflang="{lang}" href="{SITE}{url_path(code)}">'
        for code, (_, lang) in LOCALES.items()
    ) + f'\n  <link rel="alternate" hreflang="x-default" href="{SITE}">'

    for code, (name, lang) in LOCALES.items():
        assets = "" if code == "en" else "../"
        language_links = []
        for other, (label, tag) in LOCALES.items():
            href = f"{assets}{url_path(other)}" or "./"
            current = ' aria-current="page"' if code == other else ""
            language_links.append(
                f'          <li><a href="{href}" lang="{tag}" hreflang="{tag}"{current}>{escape(label)}</a></li>'
            )
        yield DOCS / url_path(code) / "index.html", template.substitute(
            **{key: escape(value) for key, value in translations[code].items()},
            site=SITE,
            repository=REPOSITORY,
            download_url=DOWNLOAD,
            canonical=f"{SITE}{url_path(code)}",
            alternates=alternates,
            language_links="\n".join(language_links),
            assets=assets,
            lang=lang,
            locale_name=escape(name),
        )

    urls = "\n".join(f"  <url><loc>{SITE}{url_path(code)}</loc></url>" for code in LOCALES)
    yield DOCS / "sitemap.xml", (
        '<?xml version="1.0" encoding="UTF-8"?>\n'
        f'<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n{urls}\n</urlset>\n'
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="Verify committed pages are up to date")
    check = parser.parse_args().check
    stale = []
    for path, content in render():
        if check:
            if not path.is_file() or path.read_text(encoding="utf-8") != content:
                stale.append(str(path.relative_to(ROOT)))
            continue
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8", newline="\n")
    if stale:
        parser.exit(1, f"Run python scripts/build-pages.py. Out of date: {', '.join(stale)}\n")


if __name__ == "__main__":
    main()
