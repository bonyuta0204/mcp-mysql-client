package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func FormatQueryResultAsJson(rows *sql.Rows) (string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get column names: %w", err)
	}

	var results = []map[string]string{}

	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		valuesPtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuesPtrs[i] = &values[i]
		}

		// Scan the row into values
		err := rows.Scan(valuesPtrs...)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings and add to result
		result := make(map[string]string)
		for i, col := range columns {
			v := values[i]
			if v.Valid {
				result[col] = v.String
			} else {
				result[col] = "NULL"
			}
		}
		results = append(results, result)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating over rows: %w", err)
	}

	json, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	return string(json), nil
}

// FormatQueryResult formats the result of a SQL query as a markdown table
func FormatQueryResult(rows *sql.Rows) (string, error) {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get column names: %w", err)
	}

	// Check if this is a SELECT query (has columns)
	if len(columns) == 0 {
		// This is likely a non-SELECT query (INSERT, UPDATE, DELETE, etc.)
		return "Query executed successfully (no results to display)", nil
	}

	// Prepare values for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Build result table
	var result string

	// Add header row
	result += "| " + columns[0]
	for i := 1; i < len(columns); i++ {
		result += " | " + columns[i]
	}
	result += " |\n"

	// Add separator row
	result += "|---"
	for i := 1; i < len(columns); i++ {
		result += "|---"
	}
	result += "|\n"

	// Add data rows
	rowCount := 0
	for rows.Next() {
		// Scan the row into values
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings and add to result
		result += "| "
		for i, val := range values {
			if i > 0 {
				result += " | "
			}

			// Handle NULL values and convert to string
			if val == nil {
				result += "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					result += string(v)
				default:
					result += fmt.Sprintf("%v", v)
				}
			}
		}
		result += " |\n"
		rowCount++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating over rows: %w", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d row(s) returned", rowCount)

	return result, nil
}

// FormatSimpleTable formats a simple table with a single column
func FormatSimpleTable(rows *sql.Rows, columnName string) (string, error) {
	// Build result
	var result string
	result += fmt.Sprintf("| %s |\n", columnName)
	result += "|----------|\n"

	var value string
	count := 0
	for rows.Next() {
		err := rows.Scan(&value)
		if err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		result += fmt.Sprintf("| %s |\n", value)
		count++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating over rows: %w", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d %s(s) found", count, columnName)

	return result, nil
}
