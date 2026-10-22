package lang

import "strings"

func EscapeMarkdownV2(s string) string {
	var replacer = strings.NewReplacer(
		`\`, `\\`,
		"_", `\_`,
		"*", `\*`,
		"[", `\[`,
		"]", `\]`,
		"(", `\(`,
		")", `\)`,
		"~", `\~`,
		">", `\>`,
		"#", `\#`,
		"+", `\+`,
		"-", `\-`,
		"=", `\=`,
		"|", `\|`,
		"{", `\{`,
		"}", `\}`,
		".", `\.`,
		"!", `\!`,
	)

	return replacer.Replace(s)
}

func Clean(s string) string {
	return strings.TrimSpace(strings.Replace(s, "`", "", -1))
}
