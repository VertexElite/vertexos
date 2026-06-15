package event

import "testing"

func TestParseSeverity(t *testing.T) {
	cases := map[string]Severity{
		"Critical":      Critical,
		"EMERGENCY":     Critical,
		"alert":         Critical,
		"Error":         Error,
		"Warning":       Warning,
		"notice":        Notice,
		"Informational": Info,
		"debug":         Info,
		"":              Info,
		"weird-value":   Notice, // unknown -> Notice
	}
	for in, want := range cases {
		if got := ParseSeverity(in); got != want {
			t.Errorf("ParseSeverity(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSeverityShortIsFixedWidth(t *testing.T) {
	for s := Info; s <= Critical; s++ {
		if len(s.Short()) != 4 {
			t.Errorf("Short(%v) = %q, want width 4", s, s.Short())
		}
	}
}
