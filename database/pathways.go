package database

import "database/sql"

// PathwayRequirementRow is a flat row from the requirement+option join --
// one row per OR-option, since a requirement can have several (e.g.
// "Either COMS W4261 or COMS E6185" produces two rows sharing a RequirementID).
type PathwayRequirementRow struct {
	RequirementID int
	GroupLabel    sql.NullString
	Title         sql.NullString
	CourseCode    sql.NullString
	RawOptionText string
}

func GetPathwayRequirements(pathwayName string) ([]PathwayRequirementRow, error) {
	rows, err := DB.Query(`
		SELECT pr.id, pr.group_label, pr.title, pro.course_code, pro.raw_option_text
		FROM pathway_requirements pr
		JOIN pathways p ON p.id = pr.pathway_id
		JOIN pathway_requirement_options pro ON pro.requirement_id = pr.id
		WHERE p.name = $1
		ORDER BY pr.id
	`, pathwayName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PathwayRequirementRow
	for rows.Next() {
		var r PathwayRequirementRow
		if err := rows.Scan(&r.RequirementID, &r.GroupLabel, &r.Title, &r.CourseCode, &r.RawOptionText); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func GetAllPathwayNames() ([]string, error) {
	rows, err := DB.Query(`SELECT name FROM pathways ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}
