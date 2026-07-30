package combine

type RecommendationCandidate struct {
	CourseCode string
	Content    string  // the embedded description text, for the LLM prompt later
	Distance   float64 // semantic relevance -- lower is more similar

	// Why this course is a legitimate recommendation, not just topically similar
	SatisfiesPathway bool
	PathwayGroup     string // e.g. "A", "B", "" -- which requirement it satisfies
	PathwayTitle     string
	SatisfiesBreadth bool
	BreadthCategory  string
}
