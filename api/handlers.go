package api

/*
	POST /recommend
	Request body:
	  {
	    "interests": "I'm interested in deep learning and neural networks",
	    "completed_courses": ["COMS 4771", "COMS 6111"],
	    "gpa": 3.5,
	    "pathway": "Machine Learning",
	    "top_k": 5            // optional, defaults to 5
	  }
	Response body:
	  {
	    "recommendations": [ ...combine.RecommendationCandidate... ],
	    "explanation": "..."   // the LLM-generated explanation
	  }
*/

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/oyeprashar/columbia_course_advisor_rag/combine"
	"github.com/oyeprashar/columbia_course_advisor_rag/generate"
)

const defaultTopK = 5

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandleRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is supported")
		return
	}

	var req recommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if req.Interests == "" {
		writeError(w, http.StatusBadRequest, "\"interests\" is required")
		return
	}
	if req.Pathway == "" {
		writeError(w, http.StatusBadRequest, "\"pathway\" is required")
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = defaultTopK
	}

	candidates, err := combine.GetRecommendations(
		req.Interests,
		req.CompletedCourses,
		req.GPA,
		req.Pathway,
		topK,
	)
	if err != nil {
		log.Printf("combine.GetRecommendations failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to compute recommendations")
		return
	}

	explanation, err := generate.GenerateRecommendation(req.Interests, candidates)
	if err != nil {
		log.Printf("generate.GenerateRecommendation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to generate explanation")
		return
	}

	writeJSON(w, http.StatusOK, recommendResponse{
		Recommendations: candidates,
		Explanation:     explanation,
	})
}
