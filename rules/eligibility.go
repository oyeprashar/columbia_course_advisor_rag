package rules

/*
	Input : completed course codes, current GPA, chosen pathway name
	Output : one combined report -- points/GPA accounting, pathway
	         completion, and breadth completion. This is what the API
	         layer calls to answer "where does this student stand."
*/

func GetEligibilityReport(completedCourses []string, studentGPA float64, pathwayName string) (*EligibilityReport, error) {
	progress, err := GetDegreeProgress(completedCourses, studentGPA)
	if err != nil {
		return nil, err
	}

	pathway, err := GetPathwayProgress(pathwayName, completedCourses)
	if err != nil {
		return nil, err
	}

	breadth, err := GetBreadthProgress(completedCourses)
	if err != nil {
		return nil, err
	}

	return &EligibilityReport{
		Progress: progress,
		Pathway:  pathway,
		Breadth:  breadth,
	}, nil
}
