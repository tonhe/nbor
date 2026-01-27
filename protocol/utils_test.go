package protocol

import "testing"

func TestCleanString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "normal string unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "removes null bytes",
			input: "hello\x00world",
			want:  "helloworld",
		},
		{
			name:  "removes multiple null bytes",
			input: "a\x00b\x00c\x00",
			want:  "abc",
		},
		{
			name:  "trims leading whitespace",
			input: "  hello",
			want:  "hello",
		},
		{
			name:  "trims trailing whitespace",
			input: "hello  ",
			want:  "hello",
		},
		{
			name:  "trims both ends",
			input: "  hello  ",
			want:  "hello",
		},
		{
			name:  "removes null and trims",
			input: "  \x00hello\x00world\x00  ",
			want:  "helloworld",
		},
		{
			name:  "handles only whitespace",
			input: "   ",
			want:  "",
		},
		{
			name:  "handles only null bytes",
			input: "\x00\x00\x00",
			want:  "",
		},
		{
			name:  "preserves internal whitespace",
			input: "hello world test",
			want:  "hello world test",
		},
		{
			name:  "handles tabs and newlines",
			input: "\thello\n",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanString(tt.input)
			if got != tt.want {
				t.Errorf("CleanString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
