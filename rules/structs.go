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

	AboveMinGPA bool
}

// OpenRequirement describes one still-unsatisfied pathway requirement --
// the student hasn't completed any of its OR'd course options yet.
type OpenRequirement struct {
	GroupLabel string
	Title      string
	Options    []string // raw option text, e.g. "Either COMS W4261 or COMS E6185"
	Codes      []string // resolved course codes only, for programmatic matching
}

type PathwayProgress struct {
	PathwayName           string
	TotalRequirements     int
	SatisfiedRequirements int
	OpenRequirements      []OpenRequirement
	IsComplete            bool
}

type BreadthProgress struct {
	TotalCategories     int
	SatisfiedCategories int
	OpenCategories      []string
	IsComplete          bool
}

// EligibilityReport combines all three checks -- this is what the API
// layer calls to answer "where does this student stand" in one shot.
type EligibilityReport struct {
	Progress *DegreeProgress
	Pathway  *PathwayProgress
	Breadth  *BreadthProgress
}
