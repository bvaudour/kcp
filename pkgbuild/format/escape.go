package format

import (
	"strings"

	"mvdan.cc/sh/v3/pattern"
)

type esc struct {
	n, d, s bool
}

var (
	escN  = esc{n: true}
	escND = esc{n: true, d: true}
	escNS = esc{n: true, s: true}

	mesc = map[rune]esc{
		' ':  escN,
		'\t': escN,
		'\n': escN,
		'|':  escN,
		'&':  escN,
		';':  escN,
		'(':  escN,
		')':  escN,
		'<':  escN,
		'>':  escN,
		'*':  escN,
		'?':  escN,
		'[':  escN,
		']':  escN,
		'~':  escN,
		'{':  escN,
		'}':  escN,
		'$':  escND,
		'`':  escND,
		'\'': escNS,
		'"':  escND,
		'#':  escN,
		'\\': escND,
	}
)

func getEsc(char rune) esc { return mesc[char] }
func needEsc(char rune, quote string) bool {
	e := getEsc(char)
	switch quote {
	case `"`:
		return e.d
	case `'`:
		return e.s
	default:
		return e.n
	}
}
func toEsc(quote string, chars ...rune) string {
	var sb strings.Builder
	for _, c := range chars {
		if needEsc(c, quote) {
			sb.WriteRune('\\')
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

func canQuote(word string) bool {
	return pattern.HasMeta(word, 0)
}
func canUnquote(word string) bool {
	for _, r := range word {
		if needEsc(r, ``) {
			return false
		}
	}
	return true
}
func canSingleQuote(word string) bool {
	for _, r := range word {
		if needEsc(r, `'`) {
			return false
		}
	}
	return true
}

func convertEsc(word, fromQuote, toQuote string) string {
	if fromQuote == toQuote {
		return word
	}

	var sb strings.Builder
	var escOn bool
	for _, r := range word {
		var buf []rune
		if escOn {
			escOn = false
			if !needEsc(r, fromQuote) {
				buf = append(buf, '\\')
			}
		} else if r == '\\' {
			escOn = true
			continue
		}
		buf = append(buf, r)
		sb.WriteString(toEsc(toQuote, buf...))
	}
	if escOn {
		sb.WriteString(toEsc(toQuote, '\\'))
	}

	return sb.String()
}

func checkConvert(word string, fromQuote string) (quotable bool, unquotable bool) {
	isQuoted := fromQuote == `'` || fromQuote == `"`
	cq, cu := canQuote(word), canUnquote(word)
	quotable = isQuoted || cq
	unquotable = (isQuoted && cu) || (!isQuoted && (!cq || cu))

	return
}

func bestConvert(word, fromQuote, toQuote string) (newWord, finalQuote string) {
	isQuotable, isUnquotable := checkConvert(word, fromQuote)
	if !isUnquotable || (toQuote != `` && isQuotable) {
		finalQuote = toQuote
		if finalQuote == `'` && !canSingleQuote(word) {
			finalQuote = `"`
		}
	}
	newWord = convertEsc(word, fromQuote, finalQuote)

	return
}
