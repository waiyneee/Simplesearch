package index

import "strings"

func Stem(token string) string {
	token = strings.TrimSpace(strings.ToLower(token))
	if token == "" {
		return token
	}
	if len(token) <= 3 {
		return token
	}

	// --- Rule 1: "-ing" ---
	// e.g. "playing" -> "play", "connecting" -> "connect"
	// Require at least 3 chars to remain after stripping.
	if strings.HasSuffix(token, "ing") && len(token) > 6 {
		stem := token[:len(token)-3]
		// Handle doubled consonant: "running" -> "runn" -> "run"
		if len(stem) >= 2 && stem[len(stem)-1] == stem[len(stem)-2] {
			stem = stem[:len(stem)-1]
		}
		return stem
	}

	// --- Rule 2: "-ed" ---
	// e.g. "played" -> "play", "connected" -> "connect"
	if strings.HasSuffix(token, "ed") && len(token) > 5 {
		stem := token[:len(token)-2]
		// Handle doubled consonant: "stopped" -> "stopp" -> "stop"
		if len(stem) >= 2 && stem[len(stem)-1] == stem[len(stem)-2] {
			stem = stem[:len(stem)-1]
		}
		return stem
	}
	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		return token[:len(token)-3] + "y"
	}
	if strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss") && len(token) > 4 {
		return token[:len(token)-1]
	}

	return token
}
