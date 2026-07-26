"""
Parses MS pathway pages (e.g. cs.columbia.edu/education/ms/machineLearning) into
structured course-list JSON.

Unlike the bulletin, these pages have no semantic CSS classes -- they're plain
<table> markup from a WYSIWYG editor. We find tables by matching header row text
("Course ID", "Title", optionally "Group") rather than by class, since there's
nothing else reliable to key off. This is inherently more fragile than the
bulletin parser -- if a page's header wording changes, the table is silently
skipped rather than mis-parsed. Check the printed skip warnings after each run.
"""

import json
import re
import sys
from pathlib import Path

from bs4 import BeautifulSoup

FOOTNOTE_RE = re.compile(r"\[\d+\]")


def clean_text(text: str) -> str:
    return text.replace("\xa0", " ").strip()


def find_header_index(header_cells, keyword):
    for i, h in enumerate(header_cells):
        if keyword in h:
            return i
    return None


def parse_pathway_html(html: str, pathway_name: str) -> list[dict]:
    soup = BeautifulSoup(html, "html.parser")
    entries = []
    tables_matched = 0

    for table in soup.find_all("table"):
        rows = table.find_all("tr")
        if not rows:
            continue

        header_cells = [clean_text(c.get_text()).lower() for c in rows[0].find_all(["td", "th"])]
        course_idx = find_header_index(header_cells, "course id")
        if course_idx is None:
            continue  # not a course-list table -- skip without complaint

        tables_matched += 1
        group_idx = find_header_index(header_cells, "group")
        title_idx = find_header_index(header_cells, "title")

        for row in rows[1:]:
            cells = row.find_all("td")
            if len(cells) <= course_idx:
                continue

            course_id_raw = clean_text(cells[course_idx].get_text(" "))
            if not course_id_raw:
                continue

            course_id_clean = FOOTNOTE_RE.sub("", course_id_raw).strip()
            course_options = [
                clean_text(opt) for opt in re.split(r"\bor\b", course_id_clean, flags=re.IGNORECASE)
            ]

            group = None
            if group_idx is not None and len(cells) > group_idx:
                group = clean_text(cells[group_idx].get_text()) or None

            title = None
            if title_idx is not None and len(cells) > title_idx:
                title = clean_text(cells[title_idx].get_text()) or None

            entries.append({
                "pathway": pathway_name,
                "group": group,
                "course_id_raw": course_id_raw,
                "course_options": course_options,
                "title": title,
            })

    if tables_matched == 0:
        print(f"  WARNING: no course-list table found in {pathway_name} "
              f"-- header wording may not match, check manually")

    return entries


# pathway file -> display name, matching data/raw/pathways/ from fetch.py
PATHWAY_FILES = {
    "computational-biology.html": "Computational Biology",
    "computer-security.html": "Computer Security",
    "foundations-of-cs.html": "Foundations of Computer Science",
    "machine-learning.html": "Machine Learning",
    "nlp.html": "Natural Language Processing",
    "network-systems.html": "Network Systems",
    "software-systems.html": "Software Systems",
    "vision-graphics-robotics.html": "Vision, Graphics, Interaction, and Robotics",
}


def run(raw_dir: str, out_path: str) -> None:
    all_entries = []

    for filename, pathway_name in PATHWAY_FILES.items():
        path = Path(raw_dir) / filename
        if not path.exists():
            print(f"  skip (not found): {path}")
            continue

        html = path.read_text()
        entries = parse_pathway_html(html, pathway_name)
        print(f"  {pathway_name}: {len(entries)} course entries")
        all_entries.extend(entries)

    Path(out_path).parent.mkdir(parents=True, exist_ok=True)
    Path(out_path).write_text(json.dumps(all_entries, indent=2))
    print(f"\nParsed {len(all_entries)} total entries across pathways -> {out_path}")


if __name__ == "__main__":
    raw_dir = sys.argv[1] if len(sys.argv) > 1 else "data/raw/pathways"
    out_path = sys.argv[2] if len(sys.argv) > 2 else "data/processed/pathways.json"
    run(raw_dir, out_path)