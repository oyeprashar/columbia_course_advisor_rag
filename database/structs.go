package database

type MSProgramRequirement struct {
	ID                  int      `db:"id"`
	ProgramName         *string  `db:"program_name"`
	TotalPointsRequired *float64 `db:"total_points_required"`
	MinimumCourseLevel  *int     `db:"minimum_course_level"`
	MinimumGPA          *float64 `db:"minimum_gpa"`
	MinPointsAt6000     *float64 `db:"min_points_at_6000_level"`
	MaxNonCSPoints      *float64 `db:"max_non_cs_points"`
	SourceURL           *string  `db:"source_url"`
}
