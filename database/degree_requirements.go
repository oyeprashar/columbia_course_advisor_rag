package database

import "log"

func GetCSmastersDegreeRequirements() *MSProgramRequirement {

	var DegreeRequirements MSProgramRequirement

	err := DB.QueryRow(`
					SELECT
						id,
						program_name,
						total_points_required,
						minimum_course_level,
						minimum_gpa,
						min_points_at_6000_level,
						max_non_cs_points,
						source_url
					FROM ms_program_requirements
					LIMIT 1
				`).Scan(
		&DegreeRequirements.ID,
		&DegreeRequirements.ProgramName,
		&DegreeRequirements.TotalPointsRequired,
		&DegreeRequirements.MinimumCourseLevel,
		&DegreeRequirements.MinimumGPA,
		&DegreeRequirements.MinPointsAt6000,
		&DegreeRequirements.MaxNonCSPoints,
		&DegreeRequirements.SourceURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	return &DegreeRequirements
}
