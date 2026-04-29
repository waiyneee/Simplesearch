package format

import "strings"

func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	lines := make([]string, 0, len(words)/4+1)
	var line strings.Builder
	curLen := 0

	for _, w := range words {
		if curLen == 0 {
			line.WriteString(w)
			curLen = len(w)
			continue
		}
		// +1 for space
		if curLen+1+len(w) > width {
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(w)
			curLen = len(w)
		} else {
			line.WriteByte(' ')
			line.WriteString(w)
			curLen += 1 + len(w)
		}
	}

	if line.Len() > 0 {
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

func TruncateLines(text string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}

	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}

	truncated := lines[:maxLines]
	last := strings.TrimSpace(truncated[len(truncated)-1])
	if last != "" {
		truncated[len(truncated)-1] = last + "..."
	} else {
		truncated[len(truncated)-1] = "..."
	}

	return strings.Join(truncated, "\n")
}
