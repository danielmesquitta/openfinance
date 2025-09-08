package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	openBrace    = '{'
	closeBrace   = '}'
	openBracket  = '['
	closeBracket = ']'
	quote        = '"'
	backslash    = '\\'
)

// ExtractJSONFromText scans the input text and returns a slice of detected JSON strings.
func ExtractJSONFromText(text string) []string {
	var jsonStrings []string
	for i := 0; i < len(text); i++ {
		if text[i] == openBrace || text[i] == openBracket {
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
	parser := &jsonParser{
		text:  text,
		start: start,
	}
	return parser.parse()
}

// jsonParser handles JSON parsing state
type jsonParser struct {
	text     string
	start    int
	stack    []byte
	inString bool
	escape   bool
}

// parse extracts JSON from the text
func (p *jsonParser) parse() (string, int, error) {
	for i := p.start; i < len(p.text); i++ {
		jsonStr, shouldReturn, err := p.processCharacter(i)
		if err != nil {
			return "", i, err
		}
		if shouldReturn {
			return jsonStr, i, nil
		}
	}

	return "", len(p.text), errors.New("no matching closing brace found")
}

// processCharacter handles a single character during parsing
func (p *jsonParser) processCharacter(pos int) (string, bool, error) {
	c := p.text[pos]

	if p.handleEscapeSequence(c) {
		return "", false, nil
	}

	if p.handleStringToggle(c) {
		return "", false, nil
	}

	if p.inString {
		return "", false, nil
	}

	if p.handleOpeningChar(c) {
		return "", false, nil
	}

	if isClosingChar(c) {
		return p.handleClosingChar(pos, c)
	}

	return "", false, nil
}

// handleEscapeSequence processes escape characters
func (p *jsonParser) handleEscapeSequence(c byte) bool {
	if p.escape {
		p.escape = false
		return true
	}

	if c == backslash {
		p.escape = true
		return true
	}

	return false
}

// handleStringToggle processes quote characters
func (p *jsonParser) handleStringToggle(c byte) bool {
	if c == quote {
		p.inString = !p.inString
		return true
	}
	return false
}

// handleOpeningChar processes opening brackets and braces
func (p *jsonParser) handleOpeningChar(c byte) bool {
	if isOpeningChar(c) {
		p.stack = append(p.stack, c)
		return true
	}
	return false
}

// handleClosingChar processes closing brackets and braces
func (p *jsonParser) handleClosingChar(pos int, c byte) (string, bool, error) {
	if len(p.stack) == 0 {
		return "", false, fmt.Errorf("mismatched '%c' at position %d", c, pos)
	}

	expectedOpening := getMatchingOpening(c)
	if p.stack[len(p.stack)-1] != expectedOpening {
		return "", false, fmt.Errorf("mismatched '%c' at position %d", c, pos)
	}

	p.stack = p.stack[:len(p.stack)-1]

	if len(p.stack) == 0 {
		jsonStr := p.text[p.start : pos+1]
		if isValidJSON(jsonStr) {
			return jsonStr, true, nil
		}
	}

	return "", false, nil
}

// isOpeningChar checks if character is an opening bracket or brace
func isOpeningChar(c byte) bool {
	return c == openBrace || c == openBracket
}

// isClosingChar checks if character is a closing bracket or brace
func isClosingChar(c byte) bool {
	return c == closeBrace || c == closeBracket
}

// getMatchingOpening returns the expected opening character for a closing character
func getMatchingOpening(closing byte) byte {
	switch closing {
	case closeBrace:
		return openBrace
	case closeBracket:
		return openBracket
	default:
		return 0
	}
}

// isValidJSON checks if a string is valid JSON.
func isValidJSON(s string) bool {
	var js any
	return json.Unmarshal([]byte(s), &js) == nil
}
