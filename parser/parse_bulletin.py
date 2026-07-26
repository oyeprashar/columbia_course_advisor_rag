"""
Parses bulletin.columbia.edu department pages into structured course JSON.

Targets the semantic markup pattern used by the bulletin's catalog platform:

    <div class="courseblock">
      <p class="courseblocktitle"><strong>CODE Title.</strong> <strong><em>points</em>.</strong></p>
      <p class="courseblockdesc"><span class="prereq">Prerequisites: ...</span> description...</p>
    </div>
"""

import json
import re
from pathlib import Path

from bs4 import BeautifulSoup

def clean_text(text: str) -> str:
    return text.replace("\xa0", " ").strip()


COURSE_CODE_RE = re.compile(r"[A-Z]{2,6}\s+[A-Z]{0,2}\d{3,4}[A-Za-z]{0,2}")
HEADER_RE = re.compile(r"^(" + COURSE_CODE_RE.pattern + r")\s+(.*)")
POINTS_RE = re.compile(r"([\d.\-]+)\s*points?")


def parse_bulletin_html(html: str) -> list[dict]:
    soup = BeautifulSoup(html, "html.parser")
    courses = []

    for block in soup.select("div.courseblock"):
        title_p = block.select_one("p.courseblocktitle")
        if not title_p:
            continue

        strongs = title_p.find_all("strong")
        if not strongs:
            continue

        header_text = clean_text(strongs[0].get_text(strip=True))
        header_match = HEADER_RE.match(header_text)
        if not header_match:
            # Doesn't match the expected "CODE Title." shape -- skip rather than
            # guess, so bad parses don't silently pollute the dataset.
            continue

        code = clean_text(header_match.group(1))
        title = clean_text(header_match.group(2)).rstrip(".")

        points = None
        if len(strongs) > 1:
            pts_match = POINTS_RE.search(clean_text(strongs[1].get_text(strip=True)))
            if pts_match:
                points = pts_match.group(1)

        desc_p = block.select_one("p.courseblockdesc")
        prereq_text = None
        description = None

        if desc_p:
            prereq_span = desc_p.select_one("span.prereq")
            if prereq_span:
                prereq_text = clean_text(prereq_span.get_text(" ", strip=True))
                prereq_span.extract()  # remove so it doesn't leak into description
            description = clean_text(desc_p.get_text(" ", strip=True)) or None

        prereq_codes = sorted(set(COURSE_CODE_RE.findall(prereq_text))) if prereq_text else []

        courses.append({
            "code": code,
            "title": title,
            "points": points,
            "prerequisites_raw": prereq_text,
            "prerequisite_codes": prereq_codes,
            "description": description,
        })

    return courses


def run(raw_path: str, out_path: str) -> None:
    html = Path(raw_path).read_text()
    courses = parse_bulletin_html(html)

    Path(out_path).parent.mkdir(parents=True, exist_ok=True)
    Path(out_path).write_text(json.dumps(courses, indent=2))

    print(f"Parsed {len(courses)} courses -> {out_path}")


if __name__ == "__main__":
    import sys

    raw_path = sys.argv[1] if len(sys.argv) > 1 else "data/raw/bulletin/cs-department.html"
    out_path = sys.argv[2] if len(sys.argv) > 2 else "data/processed/courses.json"
    run(raw_path, out_path)