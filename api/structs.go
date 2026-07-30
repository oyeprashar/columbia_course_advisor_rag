package api

import "github.com/oyeprashar/columbia_course_advisor_rag/combine"

type recommendRequest struct {
	Interests        string   `json:"interests"`
	CompletedCourses []string `json:"completed_courses"`
	GPA              float64  `json:"gpa"`
	Pathway          string   `json:"pathway"`
	TopK             int      `json:"top_k"`
}

type recommendResponse struct {
	Recommendations []combine.RecommendationCandidate `json:"recommendations"`
	Explanation     string                            `json:"explanation"`
}

type errorResponse struct {
	Error string `json:"error"`
}
