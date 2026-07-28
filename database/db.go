package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

// fall back is a local instance of postgres
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// TODO : Work on connection pooling or singleton
func getConnection() *sql.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		getEnv("PGHOST", "localhost"),
		getEnv("PGPORT", "5432"),
		getEnv("PGDATABASE", "course_advisor"),
		getEnv("PGUSER", "myuser"),
		getEnv("PGPASSWORD", "mypassword"),
	)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	return db
}

func GetCSmastersDegreeRequirements() *MSProgramRequirement {

	db := getConnection()

	var DegreeRequirements MSProgramRequirement

	err := db.QueryRow(`
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
		DegreeRequirements.ID,
		DegreeRequirements.ProgramName,
		DegreeRequirements.TotalPointsRequired,
		DegreeRequirements.MinimumCourseLevel,
		DegreeRequirements.MinimumGPA,
		DegreeRequirements.MinPointsAt6000,
		DegreeRequirements.MaxNonCSPoints,
		DegreeRequirements.SourceURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	return &DegreeRequirements
}
