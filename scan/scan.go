package scan

import (
	"strings"
	"unicode"
)

func unescaped(raw []rune, ptr int, substr string) bool {
	length := len([]rune(substr))
	if len(raw[ptr:]) < length {
		return false
	}
	if string(raw[ptr:ptr+length]) != substr {
		return false
	}
	if ptr > 0 && unescaped(raw, ptr-1, "\\") {
		return false
	}
	return true
}

func precededBy(raw []rune, ptr int, ss []string) bool {
	for _, substr := range ss {
		if ptr >= len(substr) && string(raw[ptr-len([]rune(substr)):ptr]) == substr {
			return true
		}
	}
	return false
}

func followedBy(raw []rune, ptr int, ss []string) bool {
	for _, substr := range ss {
		if len(raw[ptr+1:]) >= len(substr) && string(raw[ptr+1:ptr+1+len([]rune(substr))]) == substr {
			return true
		}
	}
	return false
}

func ScanInlineCode(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "`") && len(raw[ptr:]) > 1) {
		return 0
	}

	adv := runeIndex(string(raw[ptr+1:]), "`")
	if adv < 0 {
		return 0
	}
	cur := ptr + 1 + adv
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n"}) {
		return 0
	} else {
		return cur - ptr + 1
	}
}

func ScanInlineMath(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$") && !followedBy(raw, ptr, []string{" ", "\t"})) {
		return 0
	}

	cur := ptr + 1
	for cur < len(raw)-1 && !unescaped(raw, cur, "$") {
		adv := runeIndex(string(raw[cur+1:]), "$")
		if adv < 0 {
			return 0
		}
		cur += 1 + adv
	}
	if cur >= len(raw) {
		return 0
	}
	if precededBy(raw, cur, []string{"\n\n", "\r\n\r\n", " ", "\t"}) {
		return 0
	}
	if cur == ptr+1 {
		return 0
	}
	if unescaped(raw, cur, "$") {
		return cur - ptr + 1
	}
	return 0
}

func ScanRepeat(raw []rune, ptr int, substr string) (advance int) {
	length := len([]rune(substr))
	cur := ptr
	next := cur + length
	for len(raw[cur:]) >= length && string(raw[cur:next]) == substr {
		cur = next
		next += length
	}
	return cur - ptr
}

func ScanTag(raw []rune, ptr int) (advance int, tag string) {
	if !(unescaped(raw, ptr, "#") && len(raw[ptr:]) > 1) {
		return 0, ""
	}

	if !(unicode.IsLetter(raw[ptr+1]) || unicode.IsNumber(raw[ptr+1]) || raw[ptr+1] == '_') {
		return 0, ""
	}

	cur := ptr + 1
	for cur < len(raw) && (unicode.IsLetter(raw[cur]) || unicode.IsNumber(raw[cur]) || strings.ContainsRune("-_/", raw[cur])) {
		cur++
	}

	return cur - ptr, string(raw[ptr+1 : cur])
}

func ScanInternalLink(raw []rune, ptr int) (advance int, content string) {
	if !(unescaped(raw, ptr, "[[") && len(raw[ptr:]) >= 5) {
		return 0, ""
	}

	cur := ptr + 2
	adv := runeIndex(string(raw[cur:]), "]]")
	if adv <= 0 {
		return 0, ""
	}

	content = strings.Trim(string(raw[cur:cur+adv]), " \t")
	if strings.ContainsAny(content, "\r\n") {
		return 0, ""
	}
	cur += adv + 2
	advance = cur - ptr
	return advance, content
}

func ScanEmbeds(raw []rune, ptr int) (advance int, content string) {
	if !unescaped(raw, ptr, "![[") {
		return 0, ""
	}
	cur := ptr + 1
	advance, content = ScanInternalLink(raw, cur)
	if advance == 0 {
		return 0, ""
	}
	cur += advance
	return cur - ptr, content
}

func validURI(uri string) bool {
	return !strings.ContainsAny(uri, " \t\r\n")
}

func validExternalLinkDisplayName(displayName string) bool {
	runes := []rune(displayName)
	cur := 0
	for {
		adv := runeIndex(string(runes[cur:]), "]")
		if adv < 0 {
			return true
		}
		cur += adv // 発見した ] の位置
		if unescaped(runes, cur, "]") {
			return false
		}
		cur++
	}
}

func ScanExternalLink(raw []rune, ptr int) (advance int, displayName string, ref string) {
	if !(unescaped(raw, ptr, "[") && len(raw[ptr:]) >= 5) {
		return 0, "", ""
	}
	cur := ptr + 1 // "[" の次
	adv := runeIndex(string(raw[cur:]), "](")
	if adv < 0 || adv+1 >= len(raw[cur:])-1 {
		return 0, "", ""
	}
	next := cur + adv // "](" の "]"
	displayName = strings.Trim(string(raw[cur:next]), " \t")
	if !validExternalLinkDisplayName(displayName) {
		return 0, "", ""
	}
	cur = next
	if strings.Contains(displayName, "\r\n\r\n") || strings.Contains(displayName, "\n\n") {
		return 0, "", ""
	} else if !unescaped(raw, cur, "](") {
		return 0, "", ""
	}
	displayName = strings.Trim(displayName, "\r\n") // \r\n や \n 一つを削除
	next += 2                                       // "](" の次
	cur = next
	adv = runeIndex(string(raw[cur:]), ")")
	if adv < 0 {
		return 0, "", ""
	}
	next = cur + adv // ")"
	ref = strings.Trim(string(raw[cur:next]), " \t")
	cur = next
	if strings.Contains(ref, "\r\n\r\n") || strings.Contains(ref, "\n\n") {
		return 0, "", ""
	} else if !unescaped(raw, cur, ")") {
		return 0, "", ""
	}
	ref = strings.Trim(ref, "\r\n") // \r\n や \n 一つを削除
	if !validURI(ref) {
		return 0, "", ""
	}
	advance = next + 1 - ptr
	return
}

func ScanComment(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "%%") && len(raw) >= 2) {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "%")))
	cur := ptr + length // opening の %% の直後
	adv := runeIndex(string(raw[cur:]), strings.Repeat("%", length))
	if adv < 0 {
		return len(raw) - ptr
	}

	cur += adv + length // closing %% の直後
	return cur - ptr
}

func validMathBlockClosing(raw []rune, openPtr int, closingPtr int) bool {
	if !unescaped(raw, closingPtr, "$$") {
		return false
	}

	// 後ろに何もなければ OK
	if strings.Trim(string(raw[closingPtr+2:]), " \r\n") == "" {
		return true
	}

	// inline だったら OK
	if !strings.ContainsRune(string(raw[openPtr:closingPtr]), '\n') {
		return true
	}

	posLineFeed := runeIndex(string(raw[closingPtr+2:]), "\n")
	if posLineFeed < 0 {
		return false
	}

	remaining := string(raw[closingPtr+2 : closingPtr+2+posLineFeed])
	remaining = strings.Trim(remaining, " \r\n")
	return remaining == ""
}

func ScanMathBlock(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "$$") && len(raw) >= 3) {
		return 0
	}

	cur := ptr + 2
	for {
		adv := runeIndex(string(raw[cur:]), "$$")
		if adv < 0 {
			if raw[len(raw)-1] == '\n' {
				return len(raw) - ptr
			} else {
				return 0
			}
		}

		cur += adv
		if validMathBlockClosing(raw, ptr, cur) {
			return cur + 2 - ptr
		}
		cur += 2
	}
}

func ScanCodeBlock(raw []rune, ptr int) (advance int) {
	if !unescaped(raw, ptr, "```") {
		return 0
	}
	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "`")))
	cur := ptr + length
	adv := runeIndex(string(raw[cur:]), "```")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur += adv // closing の "```" の最初の "`"
	closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))

	// inline の場合は opening bracket と closing bracket は同じ長さでなければならない
	if !strings.ContainsRune(string(raw[ptr:cur+closingLength]), '\n') {
		if length != closingLength {
			return 0
		} else {
			return cur + closingLength - ptr
		}
	} else { // 複数行の場合は, closing bracket の長さは opening bracket の長さ以上でなければいけない
		adv := runeIndex(string(raw[cur:]), strings.Repeat("`", length))
		if adv < 0 {
			return len(raw) - ptr
		}
		cur += adv
		closingLength := len(raw[cur:]) - len([]rune(strings.TrimLeft(string(raw[cur:]), "`")))
		return cur + closingLength - ptr
	}
}

func ScanHeader(raw []rune, ptr int) (advance int, level int, headertext string) {
	if !(unescaped(raw, ptr, "#")) {
		return
	}
	// 前の改行まではスペースしか入っちゃいけない
	for back := ptr; back > 0; {
		back--
		if raw[back] == '\n' {
			break
		} else if raw[back] == ' ' {
			continue
		} else {
			return 0, 0, ""
		}
	}

	length := len(raw[ptr:]) - len([]rune(strings.TrimLeft(string(raw[ptr:]), "#")))
	cur := ptr + length // "#" の直後
	if cur >= len(raw) || !(raw[cur] == ' ' || raw[cur] == '\n' || (len(raw) >= 2 && string(raw[cur:cur+2]) == "\r\n")) {
		return 0, 0, ""
	}

	adv := runeIndex(string(raw[cur:]), "\n")
	if adv < 0 {
		advance = len(raw[ptr:])
		level = length
		headertext = strings.Trim(string(raw[cur:]), " \t\r\n")
		return advance, length, headertext
	}

	advance = cur + adv + 1 - ptr // \r\n や \n を含む
	level = length
	headertext = strings.Trim(string(raw[cur:ptr+advance]), " \t\r\n")
	return advance, level, headertext
}

func ScanEscaped(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "\\") && len(raw[ptr:]) > 1) {
		return 0
	}
	return 2
}

func ScanNormalComment(raw []rune, ptr int) (advance int) {
	if !(unescaped(raw, ptr, "<!--")) {
		return 0
	}
	adv := runeIndex(string(raw[ptr:]), "-->")
	if adv < 0 {
		return len(raw) - ptr
	}
	cur := ptr + adv + 3
	return cur - ptr
}
