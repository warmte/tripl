package lang

import "strings"

func EscapeMarkdownV2(s string) string {
	var replacer = strings.NewReplacer(
		".", "\\.",
		"!", "\\!",
		")", "\\)",
		"(", "\\(",
		"-", "\\-",
		"<", "\\<",
		">", "\\>",
		"+", "\\+",
	)

	return replacer.Replace(s)
}
