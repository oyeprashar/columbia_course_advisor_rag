package generate

/*
	Input : the student's stated interests, and the annotated candidates
	        from combine.GetRecommendations
	Output : a natural-language explanation of why each recommended course
	         fits -- grounded strictly in the candidate data, never
	         inventing courses or claiming eligibility facts not present
	         in the input

	This is the ONLY place an LLM is called in the whole pipeline. Every
	fact it's allowed to state (eligibility, pathway fit, breadth fit) was
	already determined by rules/ and retrieval/ -- the LLM's job is purely
	to explain what's already been computed, in readable prose.

	Provider is chosen via LLM_PROVIDER env var ("anthropic" or "gemini"),
	defaulting to anthropic. Both providers share the same prompt logic
	below -- only the request/response shape differs, in anthropic.go and
	gemini.go respectively.
*/

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/oyeprashar/columbia_course_advisor_rag/combine"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

const systemPrompt = `You are a course advisor for Columbia's MS in Computer Science program.

You will be given a student's stated interests and a list of candidate courses. Each candidate includes whether it satisfies a specific open pathway requirement or breadth category.

Rules you must follow strictly:
- Only discuss courses that appear in the candidate list. Never mention or invent a course that isn't listed.
- If a candidate satisfies a pathway requirement or breadth category, say so explicitly and explain which one.
- If a candidate is a general elective (satisfies neither), say that plainly rather than implying it fills a specific requirement.
- Keep the explanation grounded only in the data given to you -- do not add outside knowledge about course difficulty, workload, or instructor quality.
- Be concise: 1-2 sentences per course.`

func buildUserMessage(interests string, candidates []combine.RecommendationCandidate) string {
	var sb strings.Builder
	sb.WriteString("Student's stated interests: ")
	sb.WriteString(interests)
	sb.WriteString("\n\nCandidate courses:\n")

	for _, c := range candidates {
		sb.WriteString(fmt.Sprintf("- %s\n", c.CourseCode))
		sb.WriteString(fmt.Sprintf("  Description: %s\n", c.Content))
		if c.SatisfiesPathway {
			sb.WriteString(fmt.Sprintf("  Satisfies pathway requirement: %s (group %s)\n", c.PathwayTitle, c.PathwayGroup))
		}
		if c.SatisfiesBreadth {
			sb.WriteString(fmt.Sprintf("  Satisfies breadth category: %s\n", c.BreadthCategory))
		}
		if !c.SatisfiesPathway && !c.SatisfiesBreadth {
			sb.WriteString("  General elective (does not fill a specific pathway or breadth slot)\n")
		}
	}

	return sb.String()
}

func provider() string {
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		return strings.ToLower(v)
	}
	return "anthropic"
}

// GenerateRecommendation produces the final explained recommendation text
// from combine's already-filtered, already-annotated candidates, using
// whichever provider LLM_PROVIDER selects.
func GenerateRecommendation(interests string, candidates []combine.RecommendationCandidate) (string, error) {
	if len(candidates) == 0 {
		return "No candidate courses were found for the given interests and completed courses.", nil
	}

	userMessage := buildUserMessage(interests, candidates)

	switch provider() {
	case "gemini":
		return callGemini(systemPrompt, userMessage)
	case "anthropic":
		return callGemini(systemPrompt, userMessage) // TODO : Fix
		//return callAnthropic(systemPrompt, userMessage)
	default:
		return "", fmt.Errorf("unknown LLM_PROVIDER %q (expected \"anthropic\" or \"gemini\")", provider())
	}
}
