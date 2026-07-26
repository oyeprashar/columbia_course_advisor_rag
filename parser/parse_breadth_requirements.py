import json
from pathlib import Path

from bs4 import BeautifulSoup


def clean_text(text):
    return " ".join(text.replace("\xa0", " ").split())


def parse_breadth(html_path):

    soup = BeautifulSoup(
        Path(html_path).read_text(encoding="utf-8"),
        "html.parser",
    )

    breadth = {
        "requirements": [],
        "groups": [],
    }

    # ---------- Find heading and breadth table ----------

    heading = soup.find(
        lambda tag: tag.name in ["h1", "h2"]
                    and "Breadth Requirements" in clean_text(tag.get_text())
    )

    table = soup.find("table")

    if heading is None or table is None:
        return breadth

    # ---------- Extract requirement paragraphs ----------

    node = heading

    while True:

        node = node.find_next()

        if node is None or node == table:
            break

        if node.name == "p":

            text = clean_text(node.get_text())

            if text and text not in breadth["requirements"]:
                breadth["requirements"].append(text)

    # ---------- Extract breadth groups ----------

    rows = table.find_all("tr")

    for row in rows[1:]:

        cells = row.find_all("td")

        if len(cells) != 2:
            continue

        category = clean_text(cells[0].get_text())

        courses = [
            clean_text(text)
            for text in cells[1].stripped_strings
        ]

        breadth["groups"].append(
            {
                "category": category,
                "courses": courses,
            }
        )

    return breadth


def run(html_path, output_path):

    breadth = parse_breadth(html_path)

    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(breadth, f, indent=2)


if __name__ == "__main__":

    run(
        "data/raw/requirements/breadth-requirement.html",
        "data/processed/breadth.json",
    )