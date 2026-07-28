"""
Embeds course descriptions and FAQ Q&A pairs, writes them to the `embeddings`

    - We use all-MiniLM-L6-v2 to generate the embeddings
    - pgvector is the used and the vector column is actually treated as vectors
    - Then enables queries over the similarity


                    courses.json / faq.json
                          │
                          ▼
                    Extract text
                          │
                          ▼
                    SentenceTransformer (using the same method)
                    (all-MiniLM-L6-v2)
                          │
                          ▼
                    384-dimensional vectors
                          │
                          ▼
                    PostgreSQL (pgvector)


We are using sentence_transformers because we want sentences to have meaningful vectors that we can use
for similarity


Chunk = one document = one course_description / one faq

"""

import json
from pathlib import Path

import psycopg2
from pgvector.psycopg2 import register_vector
from sentence_transformers import SentenceTransformer

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
DATA_DIR = PROJECT_ROOT / "data" / "processed"

MODEL_NAME = "all-MiniLM-L6-v2"
BATCH_SIZE = 32


def build_course_texts(courses_path: Path) -> list[tuple[str, str, str]]:
    """Returns (source_id, source_type, content) tuples for embeddable courses.
    Skips courses with no description -- nothing meaningful to embed."""
    courses = json.loads(courses_path.read_text())
    return [
        (c["code"], "course_description", c["description"])
        for c in courses
        if c.get("description")
    ]

def build_faq_texts(faq_path: Path) -> list[tuple[str, str, str]]:
    """Combines question + answer into one embeddable text per FAQ entry,
    so semantic search matches on either the question phrasing or the
    answer content. source_id is a stable index since FAQs have no code."""
    faqs = json.loads(faq_path.read_text())
    return [
        (f"faq-{i}", "faq", f"Q: {f['question']}\nA: {f['answer']}")
        for i, f in enumerate(faqs)
    ]

def embed_and_insert(conn, model: SentenceTransformer, items: list[tuple[str, str, str]]) -> None:
    if not items:
        return

    texts = [content for _, _, content in items]
    embeddings = model.encode(texts, batch_size=BATCH_SIZE, show_progress_bar=True)

    rows = [
        (source_id, source_type, content, embedding)
        for (source_id, source_type, content), embedding in zip(items, embeddings)
    ]

    with conn.cursor() as cur:
        # Inserted one at a time (not execute_values) because the vector
        # adapter needs each embedding passed as a proper array param --
        # fine at this corpus size (~180 rows), would batch differently
        # at larger scale.
        for source_id, source_type, content, embedding in rows:
            cur.execute(
                """INSERT INTO embeddings (source_type, source_id, content, embedding)
                   VALUES (%s, %s, %s, %s)""",
                (source_type, source_id, content, embedding),
            )
    conn.commit()
    print(f"  inserted {len(rows)} rows")

def run(conn) -> None:
    print(f"Loading embedding model ({MODEL_NAME})...")
    model = SentenceTransformer(MODEL_NAME)

    print("Embedding course descriptions...")
    course_items = build_course_texts(DATA_DIR / "courses.json")
    embed_and_insert(conn, model, course_items)

    print("Embedding FAQ entries...")
    faq_items = build_faq_texts(DATA_DIR / "faq.json")
    embed_and_insert(conn, model, faq_items)

    print("\nDone.")

if __name__ == "__main__":
    from dotenv import load_dotenv

    load_dotenv(PROJECT_ROOT / ".env")

    conn = psycopg2.connect()
    register_vector(conn)  # lets psycopg2 accept numpy arrays as VECTOR params
    try:
        run(conn)
    finally:
        conn.close()