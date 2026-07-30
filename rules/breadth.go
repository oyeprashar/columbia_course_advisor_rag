package rules

/*
	Input : completed course codes
	Output : which breadth categories are covered vs. still open

	Breadth entries are either an exact course code, a wildcard pattern
	(e.g. "COMS 41xx" meaning any COMS course numbered 4100-4199), or an
	exclusion marker ("Except", "Not") -- exclusions never satisfy a
	category on their own.
*/

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	db2 "github.com/oyeprashar/columbia_course_advisor_rag/database"
)

func GetBreadthProgress(completedCourses []string) (*BreadthProgress, error) {
	entries, err := db2.GetBreadthEntries()
	if err != nil {
		return nil, err
	}

	completed := make(map[string]bool, len(completedCourses))
	for _, c := range completedCourses {
		completed[strings.ToUpper(strings.TrimSpace(c))] = true
	}

	type categoryState struct {
		satisfied bool
	}
	categories := make(map[string]*categoryState)
	var order []string

	for _, e := range entries {
		state, exists := categories[e.Category]
		if !exists {
			state = &categoryState{}
			categories[e.Category] = state
			order = append(order, e.Category)
		}

		if e.IsExclusion || state.satisfied {
			continue // exclusions never satisfy; skip further checks once satisfied
		}

		if e.CourseCode.Valid {
			if completed[strings.ToUpper(e.CourseCode.String)] {
				state.satisfied = true
			}
			continue
		}

		if e.WildcardPattern.Valid {
			re, err := wildcardToRegex(e.WildcardPattern.String)
			if err != nil {
				continue // malformed pattern in source data -- skip, don't fail the whole check
			}
			for code := range completed {
				if re.MatchString(code) {
					state.satisfied = true
					break
				}
			}
		}
	}

	progress := &BreadthProgress{}
	for _, cat := range order {
		progress.TotalCategories++
		if categories[cat].satisfied {
			progress.SatisfiedCategories++
		} else {
			progress.OpenCategories = append(progress.OpenCategories, cat)
		}
	}
	progress.IsComplete = progress.TotalCategories > 0 &&
		progress.SatisfiedCategories == progress.TotalCategories

	return progress, nil
}

// wildcardCodeRE pulls a "DEPT NNXX"-shaped code pattern out of a longer
// phrase like "All COMS 41xx courses" -- extracting just the code part
// rather than turning the whole phrase into a regex.
var wildcardCodeRE = regexp.MustCompile(`([A-Z]{2,6})\s+(\d[\dXx]{2,3})`)

func wildcardToRegex(pattern string) (*regexp.Regexp, error) {
	m := wildcardCodeRE.FindStringSubmatch(strings.ToUpper(pattern))
	if m == nil {
		return nil, fmt.Errorf("no code pattern found in %q", pattern)
	}
	dept, numPattern := m[1], m[2]

	var sb strings.Builder
	sb.WriteString("^")
	sb.WriteString(regexp.QuoteMeta(dept))
	sb.WriteString(`\s+`)
	for _, r := range numPattern {
		if r == 'X' {
			sb.WriteString(`\d`)
		} else {
			sb.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	sb.WriteString("$")
	return regexp.Compile(sb.String())
}

func nullOrEmpty(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
