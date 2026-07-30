package retrieval

/*
	Runs the actual pgvector similarity query. sourceType filters to either
	"course_description" or "faq" -- pass "" to search across both.
*/

import (
	"database/sql"
)

// Search embeds the query text and returns the topK nearest rows by cosine
// distance. db is passed in explicitly (not a package-level singleton like
// database.DB) so this package doesn't need to import database/ just to
// share the connection -- callers wire the two together in combine.go.
func Search(db *sql.DB, queryText string, sourceType string, topK int) ([]SearchResult, error) {
	queryVector, err := EmbedQuery(queryText)
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows
	if sourceType == "" {
		rows, err = db.Query(`
			SELECT source_id, source_type, content, embedding <=> $1 AS distance
			FROM embeddings
			ORDER BY embedding <=> $1
			LIMIT $2
		`, queryVector, topK)
	} else {
		rows, err = db.Query(`
			SELECT source_id, source_type, content, embedding <=> $1 AS distance
			FROM embeddings
			WHERE source_type = $2
			ORDER BY embedding <=> $1
			LIMIT $3
		`, queryVector, sourceType, topK)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.SourceID, &r.SourceType, &r.Content, &r.Distance); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
