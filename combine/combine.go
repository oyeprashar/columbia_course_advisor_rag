package combine

/*

	We have a well parsed struct telling the degree status of the student,
	and we also have relevant courses that the student can take. Now we have to
	combine the information

	Input : a student's stated interests (free text), completed courses,
	        GPA, and chosen pathway
	Output : semantically relevant courses the student hasn't taken yet,
	         each annotated with whether it also satisfies a specific open
	         pathway requirement or breadth category

	Design choice: this does NOT hard-filter to only pathway/breadth
	matches. A general elective can still count toward the 30-point
	total even if it doesn't fill a specific pathway slot -- hard
	filtering would often return zero results. Instead, only
	already-completed courses are excluded; everything else is annotated
	so the LLM step has real material to explain *why* a course fits,
	rather than just that it's topically similar.
*/

import (
	"github.com/oyeprashar/columbia_course_advisor_rag/database"
	"github.com/oyeprashar/columbia_course_advisor_rag/retrieval"
	"github.com/oyeprashar/columbia_course_advisor_rag/rules"
)

func GetRecommendations(interests string, completedCourses []string, studentGPA float64,
	pathwayName string, topK int) ([]RecommendationCandidate, error) {

	// Over-fetch semantic candidates since some will be filtered out for
	// being already completed -- otherwise a student close to done with
	// their degree could end up with fewer than topK results.
	semanticResults, err := retrieval.Search(database.DB, interests, "course_description", topK*4)
	if err != nil {
		return nil, err
	}

	report, err := rules.GetEligibilityReport(completedCourses, studentGPA, pathwayName)
	if err != nil {
		return nil, err
	}

	completed := make(map[string]bool, len(completedCourses))
	for _, c := range completedCourses {
		completed[c] = true
	}

	// Flatten open pathway requirements into code -> (group, title) so a
	// semantic result's course code can be matched back to what specific
	// requirement it would satisfy.
	type pathwayMatch struct {
		group string
		title string
	}
	pathwayByCode := make(map[string]pathwayMatch)
	for _, req := range report.Pathway.OpenRequirements {
		for _, code := range req.Codes {
			pathwayByCode[code] = pathwayMatch{group: req.GroupLabel, title: req.Title}
		}
	}

	// Same idea for breadth -- build code -> category from the raw entries.
	// Only exact-code entries are matched here (not wildcards like "COMS
	// 41xx") -- wildcard-to-candidate matching would need the same regex
	// logic as rules/breadth.go duplicated here, which isn't worth it for
	// what's ultimately just an annotation, not a hard requirement check.
	breadthEntries, err := database.GetBreadthEntries()
	if err != nil {
		return nil, err
	}
	breadthByCode := make(map[string]string)
	for _, e := range breadthEntries {
		if e.CourseCode.Valid && !e.IsExclusion {
			breadthByCode[e.CourseCode.String] = e.Category
		}
	}

	var candidates []RecommendationCandidate
	for _, result := range semanticResults {
		if completed[result.SourceID] {
			continue // already taken -- never re-recommend
		}

		candidate := RecommendationCandidate{
			CourseCode: result.SourceID,
			Content:    result.Content,
			Distance:   result.Distance,
		}

		if pm, ok := pathwayByCode[result.SourceID]; ok {
			candidate.SatisfiesPathway = true
			candidate.PathwayGroup = pm.group
			candidate.PathwayTitle = pm.title
		}

		if category, ok := breadthByCode[result.SourceID]; ok {
			candidate.SatisfiesBreadth = true
			candidate.BreadthCategory = category
		}

		candidates = append(candidates, candidate)
		if len(candidates) >= topK {
			break
		}
	}

	return candidates, nil
}
