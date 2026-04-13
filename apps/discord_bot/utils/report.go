package utils

import (
	"log"
	"regexp"
	"strings"
)

// ParseSimcReport extracts a shorter Discord-friendly summary from raw SimC
// output and prefixes it with the requesting user mention.
func ParseSimcReport(data, mentionUser string) string {
	reg, err := regexp.Compile(
		`([D|H]PS *\w+:(\n *[0-9]+\b .*)+|https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)|((\bConstant\b Buffs:)\n *(\b.*\b)*))`,
	)
	if err != nil {
		log.Printf("Failed to compile regular expression, simply returning the truncated sim data")

		return data[0:1000]
	}

	matches := reg.FindAll([]byte(data), -1)
	var sb strings.Builder
	sb.WriteString(mentionUser + "\n")
	// todo: since the regex is scuffed and captures buffs group twice,
	// todo: i am omitting the iteration over the last match with -1
	for _, match := range matches[0 : len(matches)-1] {
		sb.WriteString("\n--\n" + string(match))
	}
	final := sb.String()

	if len(final) > 1000 {
		return final[:1000]
	}

	return final
}
