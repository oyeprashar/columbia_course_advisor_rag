"""
- Read the processed jsons
- Transform
- Insert into database


├── load .env
│
├── connect to postgres
│
├── run(conn)
│      │
│      ├── load_courses()
│      │
│      ├── load_ms_requirements()
│      │
│      ├── load_pathways()
│      │
│      └── load_breadth()
│
└── close connection

"""



import json
import re
from pathlib import Path

import psycopg2
from psycopg2.extras import execute_values


PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
DATA_DIR = PROJECT_ROOT / "data" / "processed"
CODE_PARTS_RE = re.compile(r"^([A-Z]{2,6})\s+[A-Z]{0,2}(\d{3,4})[A-Za-z]{0,2}$")
WILDCARD_RE = re.compile(r"\d+x+", re.IGNORECASE)
NON_CODE_MARKERS = {"except", "not"}


def parse_points(points_str):
    """Handles both fixed credit ('3.00') and variable-credit ranges
    ('1.00-2.00', seen on topics/independent-study courses). Returns
    (min, max) -- equal for fixed-credit courses."""
    if not points_str:
        return None, None
    parts = points_str.split("-")
    try:
        if len(parts) == 2:
            return float(parts[0]), float(parts[1])
        return float(parts[0]), float(parts[0])
    except ValueError:
        return None, None


def normalize_code(code: str) -> str:
    """Canonicalize to 'DEPT NNNN' form, stripping class-level letters
    (the 'E' in 'COMS E6111') so bulletin and pathway formats match."""
    code = " ".join(code.upper().split())
    match = CODE_PARTS_RE.match(code)
    if match:
        dept, digits = match.groups()
        return f"{dept} {digits}"
    return code


def derive_level(code):
    """Course level (4000, 6000, etc.) from the numeric part of the code."""
    match = re.search(r"(\d{3,4})", code)
    if not match:
        return None
    digits = match.group(1)
    if len(digits) == 3:
        digits = "0" + digits
    return (int(digits) // 1000) * 1000


def resolve_code(raw, known_codes):
    """Returns (normalized_code_or_None, is_unresolved). A code that isn't a
    real course code at all (marker word, wildcard) resolves to (None, False)
    -- it's not 'unresolved', it was never supposed to be a code."""
    stripped = raw.strip()
    if stripped.lower() in NON_CODE_MARKERS:
        return None, False
    if WILDCARD_RE.search(stripped):
        return None, False
    norm = normalize_code(stripped)
    if norm in known_codes:
        return norm, False
    # Looked like a code but isn't in our known set -- genuinely unresolved.
    if re.match(r"^[A-Z]{2,6}\s", norm):
        return norm, True
    return None, False


# ============================================================
# Loaders -- one per source file, each owns its own transaction
# ============================================================

def load_courses(conn, path: Path) -> set[str]:
    courses = json.loads(path.read_text())
    rows = []
    for c in courses:
        code = normalize_code(c["code"])
        points_min, points_max = parse_points(c["points"])
        rows.append((code, c["title"], points_min, points_max, derive_level(code), c["description"]))

    with conn.cursor() as cur:
        execute_values(
            cur,
            """INSERT INTO courses (code, title, points_min, points_max, level, description)
               VALUES %s ON CONFLICT (code) DO NOTHING""",
            rows,
        )
    conn.commit()
    print(f"  courses: inserted {len(rows)} rows")
    return {r[0] for r in rows}


def load_ms_requirements(conn, path: Path) -> None:
    data = json.loads(path.read_text())
    with conn.cursor() as cur:
        cur.execute(
            """INSERT INTO ms_program_requirements
               (program_name, total_points_required, minimum_course_level,
                minimum_gpa, min_points_at_6000_level, max_non_cs_points, source_url)
               VALUES (%s, %s, %s, %s, %s, %s, %s)""",
            (
                data["program"],
                data["total_points_required"],
                data["minimum_course_level"],
                data["minimum_gpa"],
                data["min_points_at_6000_level"],
                data["max_non_cs_or_non_pathway_points"],
                data["source_url"],
            ),
        )
    conn.commit()
    print("  ms_program_requirements: inserted 1 row")


def load_pathways(conn, path: Path, known_codes: set[str]) -> None:
    entries = json.loads(path.read_text())
    unresolved_count = 0

    with conn.cursor() as cur:
        pathway_ids: dict[str, int] = {}

        for entry in entries:
            pathway_name = entry["pathway"]
            if pathway_name not in pathway_ids:
                cur.execute(
                    """INSERT INTO pathways (name) VALUES (%s)
                       ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
                       RETURNING id""",
                    (pathway_name,),
                )
                pathway_ids[pathway_name] = cur.fetchone()[0]

            cur.execute(
                """INSERT INTO pathway_requirements (pathway_id, group_label, title)
                   VALUES (%s, %s, %s) RETURNING id""",
                (pathway_ids[pathway_name], entry["group"], entry["title"]),
            )
            requirement_id = cur.fetchone()[0]

            for option_raw in entry["course_options"]:
                code, is_unresolved = resolve_code(option_raw, known_codes)
                if is_unresolved:
                    unresolved_count += 1
                cur.execute(
                    """INSERT INTO pathway_requirement_options
                       (requirement_id, course_code, raw_option_text, is_unresolved)
                       VALUES (%s, %s, %s, %s)""",
                    (requirement_id, code, entry["course_id_raw"], is_unresolved),
                )

    conn.commit()
    print(f"  pathways: inserted {len(entries)} requirements across "
          f"{len(pathway_ids)} pathways ({unresolved_count} unresolved options)")


def load_breadth(conn, path: Path, known_codes: set[str]) -> None:
    data = json.loads(path.read_text())
    unresolved_count = 0

    with conn.cursor() as cur:
        for group in data.get("groups", []):
            cur.execute(
                "INSERT INTO breadth_groups (category) VALUES (%s) RETURNING id",
                (group["category"],),
            )
            group_id = cur.fetchone()[0]

            for course_str in group["courses"]:
                is_exclusion = course_str.strip().lower() in NON_CODE_MARKERS
                wildcard = None
                code = None
                is_unresolved = False

                if not is_exclusion:
                    if WILDCARD_RE.search(course_str):
                        wildcard = course_str.strip()
                    else:
                        code, is_unresolved = resolve_code(course_str, known_codes)
                        if is_unresolved:
                            unresolved_count += 1

                cur.execute(
                    """INSERT INTO breadth_group_entries
                       (breadth_group_id, course_code, wildcard_pattern, is_exclusion, raw_text)
                       VALUES (%s, %s, %s, %s, %s)""",
                    (group_id, code, wildcard, is_exclusion, course_str),
                )

    conn.commit()
    print(f"  breadth: inserted groups + entries ({unresolved_count} unresolved options)")


def run(conn) -> None:
    print("Loading courses...")
    known_codes = load_courses(conn, DATA_DIR / "courses.json")

    print("Loading MS program requirements...")
    load_ms_requirements(conn, DATA_DIR / "ms_cs_requirements.json")

    print("Loading pathways...")
    load_pathways(conn, DATA_DIR / "pathways.json", known_codes)

    print("Loading breadth requirements...")
    load_breadth(conn, DATA_DIR / "breadth.json", known_codes)

    print("\nDone.")


if __name__ == "__main__":
    from dotenv import load_dotenv

    # Loads .env from the project root into the environment. psycopg2.connect()
    # with no arguments already reads PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD
    # from the environment automatically -- this just gets them there from the
    # file instead of requiring `export PGHOST=...` manually every session.
    load_dotenv(PROJECT_ROOT / ".env")

    conn = psycopg2.connect()
    try:
        run(conn)
    finally:
        conn.close()