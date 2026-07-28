-- Rule-engine schema for the Columbia CS MS course advisor.
-- Scope: CS department only for now (per current project scope).
--
-- Design principle: model AND/OR prerequisite and pathway logic as proper
-- join tables, not array/JSON columns -- so the eligibility engine can
-- express "has the student satisfied this requirement" as a plain SQL
-- query instead of application-level array parsing.

-- ============================================================
-- Courses -- the source of truth (from courses.json / bulletin)
-- ============================================================
CREATE TABLE courses (
                         code            TEXT PRIMARY KEY,        -- normalized, e.g. 'COMS E6111'
                         title           TEXT NOT NULL,
                         points_min      NUMERIC,                 -- e.g. 3.00, or 1.00 for a '1.00-2.00' range
                         points_max      NUMERIC,                 -- equals points_min for fixed-credit courses
                         level           INTEGER,                 -- 4000, 6000, etc. -- derived from code at load time
                         description     TEXT,                    -- goes to vector DB, not queried here
                         source_url      TEXT
);

-- ============================================================
-- Prerequisites -- AND of OR-groups
-- e.g. "(A or B) and (C or D)" becomes two groups, two options each
-- ============================================================
CREATE TABLE prereq_groups (
                               id              SERIAL PRIMARY KEY,
                               course_code     TEXT NOT NULL REFERENCES courses(code),
                               group_index     INTEGER NOT NULL,        -- order among AND-clauses for this course
                               raw_text        TEXT                     -- original prereq text, for display/debugging
);

CREATE TABLE prereq_options (
                                id              SERIAL PRIMARY KEY,
                                group_id        INTEGER NOT NULL REFERENCES prereq_groups(id),
                                option_code     TEXT,                    -- NULL if free-text (e.g. "instructor permission")
                                option_text     TEXT,                    -- e.g. "instructor permission" when not a course code
                                is_unresolved   BOOLEAN NOT NULL DEFAULT FALSE  -- true if option_code not found in courses table
);

-- ============================================================
-- Pathways (MS concentration tracks)
-- ============================================================
CREATE TABLE pathways (
                          id              SERIAL PRIMARY KEY,
                          name            TEXT NOT NULL UNIQUE     -- e.g. 'Machine Learning'
);

CREATE TABLE pathway_requirements (
                                      id              SERIAL PRIMARY KEY,
                                      pathway_id      INTEGER NOT NULL REFERENCES pathways(id),
                                      group_label     TEXT,                    -- e.g. 'Group A', nullable if pathway has no grouping
                                      title           TEXT                     -- descriptive title from the source table, if present
);

CREATE TABLE pathway_requirement_options (
                                             id                  SERIAL PRIMARY KEY,
                                             requirement_id      INTEGER NOT NULL REFERENCES pathway_requirements(id),
                                             course_code         TEXT,                -- normalized code, NULL if unresolved/free-text
                                             raw_option_text     TEXT NOT NULL,        -- original text, e.g. 'Either COMS W4261 or COMS E6185'
                                             is_unresolved       BOOLEAN NOT NULL DEFAULT FALSE
);

-- ============================================================
-- Breadth requirement groups
-- ============================================================
CREATE TABLE breadth_groups (
                                id              SERIAL PRIMARY KEY,
                                category        TEXT NOT NULL            -- e.g. 'Systems', 'Theory', 'AI & Applications'
);

CREATE TABLE breadth_group_entries (
                                       id                  SERIAL PRIMARY KEY,
                                       breadth_group_id    INTEGER NOT NULL REFERENCES breadth_groups(id),
                                       course_code         TEXT,                -- normalized code, NULL if wildcard/exclusion
                                       wildcard_pattern     TEXT,                -- e.g. 'COMS 41xx', NULL if a specific code
                                       is_exclusion        BOOLEAN NOT NULL DEFAULT FALSE,  -- true for entries under an "Except"/"Not" marker
                                       raw_text            TEXT NOT NULL
);

-- ============================================================
-- Global MS degree requirements (singleton config, not per-course)
-- ============================================================
CREATE TABLE ms_program_requirements (
                                         id                          SERIAL PRIMARY KEY,
                                         program_name                TEXT NOT NULL DEFAULT 'MS in Computer Science',
                                         total_points_required       NUMERIC NOT NULL,
                                         minimum_course_level        INTEGER NOT NULL,
                                         minimum_gpa                 NUMERIC NOT NULL,
                                         min_points_at_6000_level    NUMERIC,
                                         max_non_cs_points           NUMERIC,
                                         source_url                  TEXT
);

-- ============================================================
-- Indexes for the eligibility engine's actual query patterns
-- ============================================================
CREATE INDEX idx_prereq_groups_course ON prereq_groups(course_code);
CREATE INDEX idx_prereq_options_group ON prereq_options(group_id);
CREATE INDEX idx_pathway_req_pathway ON pathway_requirements(pathway_id);
CREATE INDEX idx_pathway_req_opts_req ON pathway_requirement_options(requirement_id);
CREATE INDEX idx_breadth_entries_group ON breadth_group_entries(breadth_group_id);