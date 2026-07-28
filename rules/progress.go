package rules

/*
	Input : completed course codes
	Output : student's status to complete the degree
*/
import (
	"encoding/json"
	db2 "github.com/oyeprashar/columbia_course_advisor_rag/database"
)

func GetDegressProgress([]string) *DegreeProgress {

	/*
		Steps :
			1. For the degree the user is completing, find out the requirements
	*/

	res := db2.GetCSmastersDegreeRequirements()

	marshalled, _ := json.Marshal(res)

	print(string(marshalled))

	return &DegreeProgress{}
}
