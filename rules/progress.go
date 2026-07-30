package rules

/*
	Input : completed course codes, student's current GPA
	Output : student's status against the point-based degree requirements
	         (total points, 6000-level minimum, non-CS cap, GPA)
*/

import (
	"strings"

	db2 "github.com/oyeprashar/columbia_course_advisor_rag/database"
)

// GetDegreeProgress computes points/GPA/level accounting for a student
// given their completed course codes. Pathway and breadth completion are
// handled separately in pathway.go / breadth.go, since MS prerequisites
// aren't enforced at registration -- this function only answers the
// "accounting" questions (points, level, GPA), not eligibility questions.
func GetDegreeProgress(completedCourses []string, studentGPA float64) (*DegreeProgress, error) {
	degreeRequirementsData := db2.GetCSmastersDegreeRequirements()

	courses, err := db2.GetCoursesByCodes(completedCourses)
	if err != nil {
		return nil, err
	}

	res := &DegreeProgress{
		RequiredPoints:    degreeRequirementsData.TotalPointsRequired,
		Required6000Level: degreeRequirementsData.MinPointsAt6000,
		NonCSCap:          degreeRequirementsData.MaxNonCSPoints,
		AboveMinGPA:       studentGPA >= degreeRequirementsData.MinimumGPA,
	}

	// Track which input codes were actually found in the courses table,
	// so anything left over gets reported as unrecognized rather than
	// silently ignored.
	found := make(map[string]bool, len(courses))

	for _, c := range courses {
		found[c.Code] = true

		res.TotalPoints += c.PointsMin

		if c.Level >= 6000 {
			res.Points6000Level += c.PointsMin
		}

		// Simplification for now: treat non-COMS courses as "non-CS".
		// A precise pathway-aware version of this (a course cross-listed
		// into a pathway shouldn't count against the cap) happens in
		// eligibility.go, which combines this with pathway.go's output.
		if !strings.HasPrefix(c.Code, "COMS") {
			res.NonCSOrPathwayPoints += c.PointsMin
		}
	}

	for _, code := range completedCourses {
		if !found[code] {
			res.UnrecognizedCourses = append(res.UnrecognizedCourses, code)
		}
	}

	res.PointsRemaining = res.RequiredPoints - res.TotalPoints
	if res.PointsRemaining < 0 {
		res.PointsRemaining = 0
	}

	res.Met6000Requirement = res.Points6000Level >= res.Required6000Level
	res.ExceedsNonCSCap = res.NonCSOrPathwayPoints > res.NonCSCap

	return res, nil
}
