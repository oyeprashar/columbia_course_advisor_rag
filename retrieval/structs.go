package retrieval

type SearchResult struct {
	SourceID   string
	SourceType string // "course_description" or "faq"
	Content    string
	Distance   float64 // cosine distance -- lower is more similar
}
