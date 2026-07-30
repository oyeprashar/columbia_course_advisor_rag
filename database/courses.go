package database

import (
	"github.com/lib/pq"
)

type CourseRecord struct {
	Code      string
	Title     string
	PointsMin float64
	PointsMax float64
	Level     int
}

// GetCoursesByCodes looks up known data for a set of course codes (e.g. a
// student's completed courses). Codes not found in the courses table simply
// won't appear in the result -- callers should diff the input against the
// result to detect unrecognized codes, rather than this function erroring.
func GetCoursesByCodes(codes []string) ([]CourseRecord, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	rows, err := DB.Query(`
		SELECT code, title, COALESCE(points_min, 0), COALESCE(points_max, 0), COALESCE(level, 0)
		FROM courses
		WHERE code = ANY($1)
	`, pq.Array(codes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CourseRecord
	for rows.Next() {
		var c CourseRecord
		if err := rows.Scan(&c.Code, &c.Title, &c.PointsMin, &c.PointsMax, &c.Level); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, rows.Err()
}
