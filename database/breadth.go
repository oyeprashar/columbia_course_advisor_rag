package database

import "database/sql"

type BreadthEntryRow struct {
	Category        string
	CourseCode      sql.NullString
	WildcardPattern sql.NullString
	IsExclusion     bool
	RawText         string
}

func GetBreadthEntries() ([]BreadthEntryRow, error) {
	rows, err := DB.Query(`
		SELECT bg.category, bge.course_code, bge.wildcard_pattern, bge.is_exclusion, bge.raw_text
		FROM breadth_groups bg
		JOIN breadth_group_entries bge ON bge.breadth_group_id = bg.id
		ORDER BY bg.category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BreadthEntryRow
	for rows.Next() {
		var r BreadthEntryRow
		if err := rows.Scan(&r.Category, &r.CourseCode, &r.WildcardPattern, &r.IsExclusion, &r.RawText); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
