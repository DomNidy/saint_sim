//nolint:exhaustruct
package handlers

import (
	"encoding/json"
	"testing"

	api "github.com/DomNidy/saint_sim/internal/api"
)

func TestParseAddonExport(t *testing.T) {
	t.Parallel()

	server := NewServer(stubSimulationStore{}, &stubQueue{})

	okResponse, err := server.ParseAddonExport(
		t.Context(),
		api.ParseAddonExportRequestObject{
			Body: &api.ParseAddonExportRequest{
				SimcAddonExport: "priest=\"Example\"\nlevel=80\nspec=shadow",
			},
		},
	)
	if err != nil {
		t.Fatalf("ParseAddonExport() error = %v", err)
	}

	var payload map[string]any
	okBody, marshalErr := json.Marshal(okResponse)
	if marshalErr != nil {
		t.Fatalf("marshal response: %v", marshalErr)
	}
	if err := json.Unmarshal(okBody, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	addonExport, ok := payload["addon_export"].(map[string]any)
	if !ok {
		t.Fatalf("response missing addon_export: %v", payload)
	}

	if addonExport["class"] != "priest" {
		t.Fatalf("class = %v, want priest", addonExport["class"])
	}

	badResponse, err := server.ParseAddonExport(
		t.Context(),
		api.ParseAddonExportRequestObject{
			Body: &api.ParseAddonExportRequest{
				SimcAddonExport: "### comments only",
			},
		},
	)
	if err != nil {
		t.Fatalf("ParseAddonExport() error = %v", err)
	}

	if _, ok := badResponse.(api.ParseAddonExport400JSONResponse); !ok {
		t.Fatalf(
			"response type = %T, want %T",
			badResponse,
			api.ParseAddonExport400JSONResponse{},
		)
	}
}

func TestNormalizeLineEndings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already normalized",
			input: "line1\nline2\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "windows line endings",
			input: "line1\r\nline2\r\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "classic mac line endings",
			input: "line1\rline2\rline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "mixed line endings",
			input: "line1\r\nline2\rline3\nline4",
			want:  "line1\nline2\nline3\nline4",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single trailing carriage return",
			input: "line1\r",
			want:  "line1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeLineEndings(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeLineEndings(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripAllComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Removes only commented lines",
			input: "# Simc version 1.23\nwarrior=John\nlevel=90\n",
			want:  "warrior=John\nlevel=90\n",
		},
		{
			name:  "Only commented input returns empty str",
			input: "# This is a comment line1.\n#Comment line 2\n#asd\n",
			want:  "",
		},
		{
			name:  "Multiple consecutive #'s",
			input: "##Comment #Two\n## Hey\nwarrior=John\n###123",
			want:  "warrior=John",
		},
		{
			name:  "Preserves blank lines between non-comment lines",
			input: "# Simc version 1.23\n\nwarrior=John\n\nlevel=90\n# trailing comment",
			want:  "\nwarrior=John\n\nlevel=90",
		},
		{
			name:  "Preserves spacing on non-comment lines",
			input: "# comment\n  warrior=John  \n\tlevel=90\t\n",
			want:  "  warrior=John  \n\tlevel=90\t\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := StripAllComments(tt.input)
			if got != tt.want {
				t.Fatalf("StripAllComments(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimLineWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Trims leading and trailing spaces on each line",
			input: "  warrior=John  \n  level=90  ",
			want:  "warrior=John\nlevel=90",
		},
		{
			name:  "Preserves blank lines while trimming spaces",
			input: "  warrior=John  \n   \n  level=90  ",
			want:  "warrior=John\n\nlevel=90",
		},
		{
			name:  "Normalizes line endings before trimming",
			input: "  warrior=John  \r\n  level=90  \r",
			want:  "warrior=John\nlevel=90\n",
		},
		{
			name:  "Trims tabs at the start and end of each line",
			input: "\twarrior=John\t\n \tlevel=90\t ",
			want:  "warrior=John\nlevel=90",
		},
		{
			name:  "Preserves internal tabs",
			input: "value\t=\t42\n\tname\t=\tvalue\t",
			want:  "value\t=\t42\nname\t=\tvalue",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := TrimLineWhitespace(tt.input)
			if got != tt.want {
				t.Fatalf("TrimLineWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
