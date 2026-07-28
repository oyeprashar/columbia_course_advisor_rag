package main

/*
	TODO :
		1. When the server starts we need to create the connection to the db
				or find a more efficient way
*/

import rules "github.com/oyeprashar/columbia_course_advisor_rag/rules"

func main() {

	rules.GetDegressProgress([]string{"random"})

}
