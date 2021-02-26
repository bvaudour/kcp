package runes

func IsSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func IsNL(r rune) bool {
	return r == '\n'
}

func IsBlank(r rune) bool {
	return IsSpace(r) || IsNL(r)
}

func IsEscape(r rune) bool {
	return r == '\\'
}

func IsQuote(r rune) bool {
	return r == '\'' || r == '"'
}

func IsOpen(r rune) bool {
	return r == '(' || r == '[' || r == '{'
}

func IsClose(r rune) bool {
	return r == ')' || r == ']' || r == '}'
}

func IsDelimiter(r rune) bool {
	return IsOpen(r) || IsClose(r)
}

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsLowCase(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func IsUpCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func IsLetter(r rune) bool {
	return IsLowCase(r) || IsUpCase(r)
}

func IsAlpha(r rune) bool {
	return IsLetter(r) || r == '_'
}

func IsAlphaNum(r rune) bool {
	return IsAlpha(r) || IsDigit(r)
}

func IsComment(r rune) bool {
	return r == '#'
}

func IsVariable(r rune) bool {
	return r == '$'
}

func IsAffectation(r rune) bool {
	return r == '='
}

//IsEscapable returns true if the given
//rune should be escaped according to the given quote.
func IsEscapable(r rune, quote rune) bool {
	if IsEscape(r) {
		return true
	}
	switch quote {
	case '\'':
		return r == quote
	case '"':
		return r == quote || IsVariable(r) || IsDelimiter(r)
	}
	return IsQuote(r) || IsBlank(r) || IsDelimiter(r) || IsComment(r) || IsVariable(r)
}

//CheckerFunc is a function which checks a rune.
type CheckerFunc func(rune) bool

//BlockCheckerFunc is a function which checks/blocks a rune.
type BlockCheckerFunc func(rune) (bool, error)

//ReverseChecker returns the negative checker
//of the given callback.
func ReverseChecker(cb CheckerFunc) CheckerFunc {
	return func(r rune) bool { return !cb(r) }
}

func Checker2Block(cb CheckerFunc) BlockCheckerFunc {
	return func(r rune) (bool, error) { return cb(r), nil }
}

func RChecker2Block(cb CheckerFunc) BlockCheckerFunc {
	return Checker2Block(ReverseChecker(cb))
}

//CheckString checks if all runes of the string
//pass the given checker func.
func CheckString(s string, cb CheckerFunc) bool {
	for _, r := range s {
		if !cb(r) {
			return false
		}
	}
	return true
}
