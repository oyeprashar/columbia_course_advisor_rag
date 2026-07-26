"""
We want to download all the relevant HTMLS
"""

import json
import time
import urllib.robotparser
from dataclasses import dataclass
from pathlib import Path
from urllib.parse import urlparse
import requests

PROJECT_ROOT = Path(__file__).resolve().parent.parent
RAW_DIR = PROJECT_ROOT / "data" / "raw"
MANIFEST_PATH = RAW_DIR / "manifest.json"

HEADERS = {
    "User-Agent": "course-rag-project/0.1 (personal academic project; contact: sp4540@columbia.edu)"
}

RATE_LIMIT_SECONDS = 1.5
MAX_RETRIES = 3
RETRY_BACKOFF_BASE = 2  # seconds, doubles each retry


@dataclass
class Target:
    url: str
    save_path: str  # relative to data/raw/


# Explicit target list -- not a crawler. Add pages deliberately.
TARGETS = [
    # Course descriptions, prerequisites, live scheduling
    Target(
        url="https://bulletin.columbia.edu/columbia-engineering/academic-departments-programs/computer-science/",
        save_path="bulletin/cs-department.html",
    ),

    # MS degree requirements hub
    Target(
        url="https://www.cs.columbia.edu/education/ms/",
        save_path="requirements/ms-cs-requirements.html",
    ),

    # What counts toward breadth
    Target(
        url="https://www.cs.columbia.edu/education/ms/breadthRequirement",
        save_path="requirements/breadth-requirement.html",
    ),

    # Registration/eligibility nuance (prereqs not enforced, transfer credit, etc.)
    Target(
        url="https://www.cs.columbia.edu/education/ms/regfaq/",
        save_path="requirements/ms-faq.html",
    ),

    # Pathway course lists (excludes MS Personalized / MS Thesis -- invite-only, not general)
    Target(
        url="https://www.cs.columbia.edu/education/ms/computationalBiology",
        save_path="pathways/computational-biology.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/newComputerSecurity",
        save_path="pathways/computer-security.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/foundationsOfCS",
        save_path="pathways/foundations-of-cs.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/machineLearning",
        save_path="pathways/machine-learning.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/nlp",
        save_path="pathways/nlp.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/networkSystems",
        save_path="pathways/network-systems.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/softwareSystems",
        save_path="pathways/software-systems.html",
    ),
    Target(
        url="https://www.cs.columbia.edu/education/ms/visionAndGraphics",
        save_path="pathways/vision-graphics-robotics.html",
    ),
]


def is_allowed_by_robots(url: str) -> bool:
    parsed = urlparse(url)
    robots_url = f"{parsed.scheme}://{parsed.netloc}/robots.txt"
    rp = urllib.robotparser.RobotFileParser()
    try:
        rp.set_url(robots_url)
        rp.read()
    except Exception:
        # If robots.txt is unreachable, fail closed: don't assume permission.
        print(f"  could not read {robots_url}; skipping to be safe")
        return False
    return rp.can_fetch(HEADERS["User-Agent"], url)


def fetch_with_retry(url):
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            resp = requests.get(url, headers=HEADERS, timeout=15)
            resp.raise_for_status()
            return resp.text
        except requests.RequestException as e:
            wait = RETRY_BACKOFF_BASE ** attempt
            print(f"  attempt {attempt}/{MAX_RETRIES} failed ({e}); retrying in {wait}s")
            time.sleep(wait)
    print(f"  giving up on {url} after {MAX_RETRIES} attempts")
    return None


def load_manifest() -> dict:
    if MANIFEST_PATH.exists():
        return json.loads(MANIFEST_PATH.read_text())
    return {}


def save_manifest(manifest: dict) -> None:
    MANIFEST_PATH.parent.mkdir(parents=True, exist_ok=True)
    MANIFEST_PATH.write_text(json.dumps(manifest, indent=2))


def run(force: bool = False) -> None:
    manifest = load_manifest()

    for target in TARGETS:
        out_path = RAW_DIR / target.save_path

        if out_path.exists() and not force:
            print(f"skip (already fetched): {target.url}")
            continue

        print(f"fetching: {target.url}")

        if not is_allowed_by_robots(target.url):
            print(f"  blocked by robots.txt, skipping: {target.url}")
            continue

        html = fetch_with_retry(target.url)
        if html is None:
            continue

        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(html)

        manifest[target.save_path] = {
            "url": target.url,
            "fetched_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }
        save_manifest(manifest)

        time.sleep(RATE_LIMIT_SECONDS)

    print(f"\ndone. {len(manifest)} pages tracked in manifest.")


if __name__ == "__main__":
    import sys

    force_flag = "--force" in sys.argv
    run(force=force_flag)