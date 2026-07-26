import json
from pathlib import Path

from bs4 import BeautifulSoup


def clean_text(text):
    return " ".join(text.replace("\xa0", " ").split())


def parse_faq(html_path):

    soup = BeautifulSoup(
        Path(html_path).read_text(encoding="utf-8"),
        "html.parser",
    )

    faqs = []

    panels = soup.find_all("div", class_="panel panel-default")

    for panel in panels:

        question_tag = panel.find("h4", class_="panel-title")
        answer_tag = panel.find("div", class_="panel-body")

        if not question_tag or not answer_tag:
            continue

        question = clean_text(question_tag.get_text())
        answer = clean_text(answer_tag.get_text(" "))

        faqs.append(
            {
                "question": question,
                "answer": answer,
            }
        )

    return faqs


def run(html_path, output_path):

    faqs = parse_faq(html_path)

    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(faqs, f, indent=2)


if __name__ == "__main__":

    run(
        "data/raw/requirements/ms-faq.html",
        "data/processed/faq.json",
    )