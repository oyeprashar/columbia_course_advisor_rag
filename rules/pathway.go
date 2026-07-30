package rules

/*
	Input : a chosen pathway name, completed course codes
	Output : which pathway requirements are satisfied vs. still open

	Each pathway requirement can have multiple OR'd course options (e.g.
	"Either COMS W4261 or COMS E6185") -- satisfied if at least one option
	is in the student's completed courses.
*/

import (
	db2 "github.com/oyeprashar/columbia_course_advisor_rag/database"
)

func GetPathwayProgress(pathwayName string, completedCourses []string) (*PathwayProgress, error) {
	rows, err := db2.GetPathwayRequirements(pathwayName)
	if err != nil {
		return nil, err
	}

	completed := make(map[string]bool, len(completedCourses))
	for _, c := range completedCourses {
		completed[c] = true
	}

	// Group the flat rows by requirement ID -- each requirement may span
	// several rows (the OR-options), all sharing one RequirementID.
	type requirementGroup struct {
		groupLabel string
		title      string
		options    []string // raw text, for display when a requirement is open
		codes      []string // resolved codes only, for the satisfied check
	}
	groups := make(map[int]*requirementGroup)
	var order []int

	for _, row := range rows {
		g, exists := groups[row.RequirementID]
		if !exists {
			g = &requirementGroup{
				groupLabel: nullOrEmpty(row.GroupLabel),
				title:      nullOrEmpty(row.Title),
			}
			groups[row.RequirementID] = g
			order = append(order, row.RequirementID)
		}
		g.options = append(g.options, row.RawOptionText)
		if row.CourseCode.Valid {
			g.codes = append(g.codes, row.CourseCode.String)
		}
	}

	progress := &PathwayProgress{PathwayName: pathwayName}

	for _, id := range order {
		g := groups[id]
		progress.TotalRequirements++

		satisfied := false
		for _, code := range g.codes {
			if completed[code] {
				satisfied = true
				break
			}
		}

		if satisfied {
			progress.SatisfiedRequirements++
		} else {
			progress.OpenRequirements = append(progress.OpenRequirements, OpenRequirement{
				GroupLabel: g.groupLabel,
				Title:      g.title,
				Options:    g.options,
				Codes:      g.codes,
			})
		}
	}

	progress.IsComplete = progress.TotalRequirements > 0 &&
		progress.SatisfiedRequirements == progress.TotalRequirements

	return progress, nil
}
