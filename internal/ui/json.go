package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const maxDisplayLines = 3000

func HighlightJSON(data []byte, styles Styles, withLineNumbers bool) string {
	if len(data) == 0 {
		return styles.FooterDesc.Render("  (Empty response body)")
	}

	var rawStr string
	var v any
	if err := json.Unmarshal(data, &v); err == nil {
		var prettyJSON bytes.Buffer
		prettyJSON.Grow(len(data) + (len(data) / 2))
		enc := json.NewEncoder(&prettyJSON)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err == nil {
			rawStr = strings.TrimRight(prettyJSON.String(), "\n")
		} else {
			rawStr = string(data)
		}
	} else {
		rawStr = string(data)
	}

	lines := strings.Split(rawStr, "\n")
	totalLines := len(lines)
	truncated := false
	if totalLines > maxDisplayLines {
		lines = lines[:maxDisplayLines]
		truncated = true
	}

	maxDigits := len(fmt.Sprintf("%d", totalLines))
	if maxDigits < 2 {
		maxDigits = 2
	}

	var out strings.Builder
	out.Grow(len(rawStr) * 2)

	for i, line := range lines {
		colored := highlightJSONLine(line, styles)

		if withLineNumbers {
			num := fmt.Sprintf("%*d", maxDigits, i+1)
			out.WriteString(styles.GutterNumber.Render(num))
			out.WriteString(styles.GutterSep.Render(" │ "))
			out.WriteString(colored)
		} else {
			out.WriteString("  ")
			out.WriteString(colored)
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}

	if truncated {
		out.WriteString("\n\n  ")
		out.WriteString(styles.FooterDesc.Render(fmt.Sprintf("... [Truncated %d lines for performance]", totalLines-maxDisplayLines)))
	}

	return out.String()
}

func highlightJSONLine(line string, styles Styles) string {
	n := len(line)
	if n == 0 {
		return ""
	}

	var out strings.Builder
	out.Grow(n + 32)
	i := 0

	for i < n && (line[i] == ' ' || line[i] == '\t') {
		out.WriteByte(line[i])
		i++
	}

	if i < n && line[i] == '"' {
		start := i
		i++
		escaped := false
		for i < n {
			if line[i] == '\\' && !escaped {
				escaped = true
				i++
				continue
			}
			if line[i] == '"' && !escaped {
				i++
				break
			}
			escaped = false
			i++
		}
		candidate := line[start:i]

		j := i
		for j < n && (line[j] == ' ' || line[j] == '\t') {
			j++
		}

		if j < n && line[j] == ':' {
			if len(candidate) >= 2 {
				keyName := candidate[1 : len(candidate)-1]
				out.WriteString(fmt.Sprintf(`"%s"`, styles.JsonKey.Render(keyName)))
			} else {
				out.WriteString(candidate)
			}

			for i < j {
				out.WriteByte(line[i])
				i++
			}
			out.WriteString(styles.JsonBrace.Render(":"))
			i++

			if i < n && line[i] == ' ' {
				out.WriteByte(' ')
				i++
			}
		} else {
			i = start
		}
	}

	for i < n {
		b := line[i]

		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			out.WriteByte(b)
			i++
			continue
		}

		if b == '{' || b == '}' || b == '[' || b == ']' || b == ',' {
			out.WriteString(styles.JsonBrace.Render(string(b)))
			i++
			continue
		}

		if b == '"' {
			start := i
			i++
			escaped := false
			for i < n {
				if line[i] == '\\' && !escaped {
					escaped = true
					i++
					continue
				}
				if line[i] == '"' && !escaped {
					i++
					break
				}
				escaped = false
				i++
			}
			strVal := line[start:i]
			if len(strVal) >= 2 {
				inner := strVal[1 : len(strVal)-1]
				out.WriteString(fmt.Sprintf(`"%s"`, styles.JsonString.Render(inner)))
			} else {
				out.WriteString(strVal)
			}
			continue
		}

		if (b >= '0' && b <= '9') || b == '-' {
			start := i
			for i < n && ((line[i] >= '0' && line[i] <= '9') || line[i] == '.' || line[i] == 'e' || line[i] == 'E' || line[i] == '+' || line[i] == '-') {
				i++
			}
			numVal := line[start:i]
			out.WriteString(styles.JsonNumber.Render(numVal))
			continue
		}

		if unicode.IsLetter(rune(b)) {
			start := i
			for i < n && unicode.IsLetter(rune(line[i])) {
				i++
			}
			word := line[start:i]
			switch word {
			case "true", "false":
				out.WriteString(styles.JsonBool.Render(word))
			case "null":
				out.WriteString(styles.JsonNull.Render(word))
			default:
				out.WriteString(word)
			}
			continue
		}

		out.WriteByte(b)
		i++
	}

	return out.String()
}
