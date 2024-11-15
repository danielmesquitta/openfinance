package jsonutil

import (
	"encoding/json"
	"fmt"
)

// ExtractJSONFromText scans the input text and returns a slice of detected JSON strings.
func ExtractJSONFromText(text string) []string {
	var jsonStrings []string
	for i := 0; i < len(text); i++ {
		if text[i] == '{' || text[i] == '[' {
			jsonStr, endPos, err := extractJSON(text, i)
			if err == nil {
				jsonStrings = append(jsonStrings, jsonStr)
				i = endPos // Move index to the end of the detected JSON to prevent overlapping
			}
		}
	}
	return jsonStrings
}

// extractJSON attempts to extract a JSON string starting from the given position.
func extractJSON(text string, start int) (string, int, error) {
	var stack []byte
	inString := false
	escape := false

	for i := start; i < len(text); i++ {
		c := text[i]

		if escape {
			escape = false
			continue
		}
		if c == '\\' {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
		}

		if !inString {
			if c == '{' || c == '[' {
				stack = append(stack, c)
			} else if c == '}' {
				if len(stack) == 0 || stack[len(stack)-1] != '{' {
					return "", i, fmt.Errorf("mismatched '}' at position %d", i)
				}
				stack = stack[:len(stack)-1]
				if len(stack) == 0 {
					jsonStr := text[start : i+1]
					if isValidJSON(jsonStr) {
						return jsonStr, i, nil
					}
				}
			} else if c == ']' {
				if len(stack) == 0 || stack[len(stack)-1] != '[' {
					return "", i, fmt.Errorf("mismatched ']' at position %d", i)
				}
				stack = stack[:len(stack)-1]
				if len(stack) == 0 {
					jsonStr := text[start : i+1]
					if isValidJSON(jsonStr) {
						return jsonStr, i, nil
					}
				}
			}
		}
	}
	return "", len(text), fmt.Errorf("no matching closing brace found")
}

// isValidJSON checks if a string is valid JSON.
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
