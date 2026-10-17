package transpiler

import (
	"errors"
	"unicode"
)

// Read declaration tokens, ignoring comments and literals that can contain code
// examples. Keep the object name intact, including Scala's backtick identifiers.
func applicationEntry(code string) (string, error) {
	tokens := scalaTokens([]rune(code))
	var entry string
	for i := 0; i+3 < len(tokens); i++ {
		if tokens[i] != "object" || tokens[i+2] != "extends" {
			continue
		}
		base := i + 3
		if tokens[base] == "zio" && base+2 < len(tokens) && tokens[base+1] == "." {
			base += 2
		}
		if tokens[base] != "ZIOAppDefault" {
			continue
		}
		if entry != "" {
			return "", errors.New("ZIOAppDefault を継承する object は1つにしてください。")
		}
		entry = tokens[i+1]
	}
	if entry == "" {
		return "", errors.New("ZIOAppDefault を継承する object と、その run を定義してください。")
	}
	return entry, nil
}

func scalaTokens(code []rune) []string {
	var tokens []string
	for i := 0; i < len(code); {
		switch {
		case unicode.IsSpace(code[i]):
			i++
		case hasRunes(code, i, "//"):
			for i < len(code) && code[i] != '\n' {
				i++
			}
		case hasRunes(code, i, "/*"):
			i += 2
			for depth := 1; i < len(code) && depth > 0; {
				switch {
				case hasRunes(code, i, "/*"):
					depth++
					i += 2
				case hasRunes(code, i, "*/"):
					depth--
					i += 2
				default:
					i++
				}
			}
		case hasRunes(code, i, `"""`):
			i += 3
			for i < len(code) && !hasRunes(code, i, `"""`) {
				i++
			}
			i += min(3, len(code)-i)
		case code[i] == '\'' && i+2 < len(code) && code[i+2] == '\'':
			i += 3
		case code[i] == '\'' && i+3 < len(code) && code[i+1] == '\\' && code[i+3] == '\'':
			i += 4
		case code[i] == '"':
			i++
			for i < len(code) {
				if code[i] == '\\' {
					i += min(2, len(code)-i)
					continue
				}
				if code[i] == '"' {
					i++
					break
				}
				i++
			}
		case code[i] == '`':
			start := i
			i++
			for i < len(code) && code[i] != '`' {
				i++
			}
			i += min(1, len(code)-i)
			tokens = append(tokens, string(code[start:i]))
		case unicode.IsLetter(code[i]) || code[i] == '_' || code[i] == '$':
			start := i
			i++
			for i < len(code) && (unicode.IsLetter(code[i]) || unicode.IsDigit(code[i]) || code[i] == '_' || code[i] == '$') {
				i++
			}
			tokens = append(tokens, string(code[start:i]))
		default:
			tokens = append(tokens, string(code[i]))
			i++
		}
	}
	return tokens
}

func hasRunes(code []rune, at int, value string) bool {
	chars := []rune(value)
	if len(code)-at < len(chars) {
		return false
	}
	for i, ch := range chars {
		if code[at+i] != ch {
			return false
		}
	}
	return true
}
