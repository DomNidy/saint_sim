// Package simc provides various utilities for performing transformations
// on TCI strings.
//
// TCI (Textual Configuration Interface) is the language used to configure
// simc.
//
// TCI Docs:
//   - https://github.com/simulationcraft/simc/wiki/TextualConfigurationInterface
package simc

import "strings"

// NormalizeLineEndings converts CRLF (\r\n) and CR (\r) line endings to LF (\n).
func NormalizeLineEndings(tciString string) string {
	// Normalize CRLF first so the subsequent CR replacement does not
	// introduce double newlines.
	return strings.ReplaceAll(
		strings.ReplaceAll(tciString, "\r\n", "\n"),
		"\r",
		"\n",
	)
}

// StripAllComments returns a new string where each commented line in the
// TCI profile is removed.
//
// TCI comments are lines that begin with a "#".
// TCI does not recognize trailing comments, so we don't try to remove them.
func StripAllComments(tciString string) string {
	tciString = NormalizeLineEndings(tciString)
	lines := strings.Split(tciString, "\n")

	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// TrimLineWhitespace trims leading and trailing spaces and tabs from each
// line in the TCI string.
func TrimLineWhitespace(tciString string) string {
	tciString = NormalizeLineEndings(tciString)
	lines := strings.Split(tciString, "\n")

	trimmedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLines = append(trimmedLines, strings.Trim(line, " \t"))
	}

	return strings.Join(trimmedLines, "\n")
}
