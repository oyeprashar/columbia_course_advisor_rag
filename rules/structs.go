package rules

type DegreeProgress struct {
	TotalPoints     float64 // sum of points_min across all completed courses
	RequiredPoints  float64 // 30, from ms_program_requirements
	PointsRemaining float64 // RequiredPoints - TotalPoints (0 if met/exceeded)

	Points6000Level    float64 // sum of points where level >= 6000
	Required6000Level  float64 // 6, from ms_program_requirements
	Met6000Requirement bool

	NonCSOrPathwayPoints float64 // points from courses outside CS/pathway scope
	NonCSCap             float64 // 3, from ms_program_requirements
	ExceedsNonCSCap      bool

	UnrecognizedCourses []string // codes passed in that weren't found in the courses table
}
