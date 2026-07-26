import json
from pathlib import Path

from bs4 import BeautifulSoup


def clean_text(text):
    return " ".join(text.replace("\xa0", " ").split())


def parse_requirements(html_path):

    soup = BeautifulSoup(
        Path(html_path).read_text(encoding="utf-8"),
        "html.parser",
    )

    requirements = []

    content_root = soup.select_one("article") or soup

    # Parse every H2 section
    for heading in content_root.find_all("h2"):

        title = clean_text(heading.get_text())

        content = []

        curr = heading.find_next_sibling()

        while curr and curr.name != "h2":

            # Paragraphs
            if curr.name == "p":
                text = clean_text(curr.get_text())
                if text:
                    content.append(text)

            # Bullet lists
            elif curr.name == "ul":
                for li in curr.find_all("li", recursive=False):
                    text = clean_text(li.get_text())
                    if text:
                        content.append(text)

            # Tables
            elif curr.name == "table":
                rows = []

                for tr in curr.find_all("tr"):
                    cells = [
                        clean_text(td.get_text())
                        for td in tr.find_all(["th", "td"])
                    ]

                    if cells:
                        rows.append(cells)

                if rows:
                    content.append(rows)

            curr = curr.find_next_sibling()

        requirements.append(
            {
                "section": title,
                "content": content,
            }
        )

    return requirements


def run(html_path, output_path):

    requirements = parse_requirements(html_path)

    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(requirements, f, indent=2)


if __name__ == "__main__":

    run(
        "data/raw/requirements/ms-cs-requirements.html",
        "data/processed/requirements.json",
    )