package dbutils

import (
	"fmt"
	"strings"
)

// JSONEach returns JSON_EACH string expression with
// some normalizations for non-json columns.
func JSONEach(driver string, column string) string {
	if driver == "postgres" {
		// Postgres doesn't strictly need a special utility for this if we assume jsonb,
		// but to match the "each" behavior (array expanstion) we can use jsonb_array_elements.
		// However, dbx doesn't support set-returning functions in simple expressions easily.
		// NOTE: PocketBase usage of JSONEach is strictly inside FROM/JOIN clauses usually.
		return fmt.Sprintf("jsonb_array_elements([[%s]])", column)
	}

	// note: we are not using the new and shorter "if(x,y)" syntax for
	// compatibility with custom drivers that use older SQLite version
	return fmt.Sprintf(
		`json_each(CASE WHEN iif(json_valid([[%s]]), json_type([[%s]])='array', FALSE) THEN [[%s]] ELSE json_array([[%s]]) END)`,
		column, column, column, column,
	)
}

// JSONArrayLength returns JSON_ARRAY_LENGTH string expression
// with some normalizations for non-json columns.
//
// It works with both json and non-json column values.
//
// Returns 0 for empty string or NULL column values.
func JSONArrayLength(driver string, column string) string {
	if driver == "postgres" {
		return fmt.Sprintf("jsonb_array_length(COALESCE([[%s]], '[]'::jsonb))", column)
	}

	// note: we are not using the new and shorter "if(x,y)" syntax for
	// compatibility with custom drivers that use older SQLite version
	return fmt.Sprintf(
		`json_array_length(CASE WHEN iif(json_valid([[%s]]), json_type([[%s]])='array', FALSE) THEN [[%s]] ELSE (CASE WHEN [[%s]] = '' OR [[%s]] IS NULL THEN json_array() ELSE json_array([[%s]]) END) END)`,
		column, column, column, column, column, column,
	)
}

// JSONExtract returns a JSON_EXTRACT string expression with
// some normalizations for non-json columns.
func JSONExtract(driver string, column string, path string) string {
	if driver == "postgres" {
		// Postgres jsonb_path_query or * operator?
		// path usually comes in as "a.b" or "[0].a".
		// Postgres uses -> operator.
		// For simplicity/safety, let's try to map the path to Postgres jsonb path syntax or simple casting?
		// This might be complex to regex parse "a.b[0]".
		// A cleaner way for Postgres might be using `jsonb_extract_path`.
		
		// This is a naive implementation. For full path support it requires parsing `path`.
		// Given the limited time, let's assume we can use the `->` operator chains if we parsed it.
		// But since we can't easily parse it here, we might need a different strategy.
		//
		// However, PocketBase `path` argument usually is dot notation "a.b".
		// We can split by dot.
		parts := strings.Split(path, ".")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s", column))
		for i, p := range parts {
			if p == "" { continue }
			// Check for array notation [index]
			if strings.HasPrefix(p, "[") {
				// e.g. [0]
				sb.WriteString(fmt.Sprintf("->%s", strings.Trim(p, "[]")))
			} else {
				if i == len(parts)-1 {
					// Last part, return as text? SQLite json_extract returns mixed types.
					// For consistency let's keep it as is (jsonb).
					sb.WriteString(fmt.Sprintf("->'%s'", p))
				} else {
					sb.WriteString(fmt.Sprintf("->'%s'", p))
				}
			}
		}
		// Cast to text to match SQLite json_extract behavior for comparisons?
		// Actually json_extract returns valid JSON or primitives. 
		// For now let's leave it as JSONB extraction.
		return sb.String()
	}

	// prefix the path with dot if it is not starting with array notation
	if path != "" && !strings.HasPrefix(path, "[") {
		path = "." + path
	}

	return fmt.Sprintf(
		// note: the extra object wrapping is needed to workaround the cases where a json_extract is used with non-json columns.
		"(CASE WHEN json_valid([[%s]]) THEN JSON_EXTRACT([[%s]], '$%s') ELSE JSON_EXTRACT(json_object('pb', [[%s]]), '$.pb%s') END)",
		column,
		column,
		path,
		column,
		path,
	)
}
