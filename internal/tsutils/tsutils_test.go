package tsutils

import "testing"

func TestMentionsDoubleAtTag(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"", false},
		{"@@hidden", true},
		{"text @@hidden more", true},
		{"@@hidden.", true},
		{"@@@hidden", true},
		{"@hidden", false},
		{"@@hiddenX", false},
		{"@@hidden-x", false},
		{"@@hidden_x", false},
		{"@@hiddenX @@hidden", true},
		{"@@hiddenX @@hidden2", false},
	}
	for _, tt := range tests {
		if got := mentionsDoubleAtTag(tt.text, "hidden"); got != tt.want {
			t.Errorf("mentionsDoubleAtTag(%q, \"hidden\") = %v, want %v", tt.text, got, tt.want)
		}
	}
}
