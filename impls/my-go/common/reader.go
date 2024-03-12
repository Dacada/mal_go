package common;

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"errors"
)

type tokenizer struct {
	Input string
	HasPeeked bool
	Peeked string
	HasPrevChar bool
	PrevChar rune
}

func newTokenizer(input string) tokenizer {
	return tokenizer{input, false, "", false, 0}
}

func (t *tokenizer) unadvanceChar(c rune) {
	t.HasPrevChar = true
	t.PrevChar = c
}

func (t *tokenizer) advanceChar() (rune, error) {
	if t.HasPrevChar {
		t.HasPrevChar = false;
		return t.PrevChar, nil
	}
	
	r, size := utf8.DecodeRuneInString(t.Input)
	if r == utf8.RuneError {
		return 0, errors.New("invalid utf8 input")
	}
	t.Input = t.Input[size:]
	return r, nil
}

func (t *tokenizer) peek() (string, error) {
	if !t.HasPeeked {
		token, err := t.next()
		if err != nil {
			return "", err
		}
		t.Peeked = token
		t.HasPeeked = true
	}
	return t.Peeked, nil
}

func (t *tokenizer) next() (string, error) {
	if t.HasPeeked {
		t.HasPeeked = false
		return t.Peeked, nil
	}

	var curr rune
	if t.HasPrevChar {
		curr = t.PrevChar
		t.HasPrevChar = false
	} else {
		if t.Input == "" {
			return "", nil
		}
		r, err := t.advanceChar()
		curr = r
		if err != nil {
			return "", err
		}
	}

	for {
		if curr != ',' && !unicode.IsSpace(curr) {
			break;
		}

		if t.Input == "" {
			return "", nil
		}
		r, err := t.advanceChar()
		curr = r
		if err != nil {
			return "", err
		}
	}

	if curr == '~' {
		if t.Input == "" {
			return "~", nil
		}
		r, err := t.advanceChar()
		if err != nil {
			return "", err
		}
		if r == '@' {
			return "~@", nil
		} else {
			t.unadvanceChar(r)
			return "~", nil
		}
	}

	if curr == '[' || curr == ']' || curr == '{' || curr == '}' || curr == '(' || curr == ')' || curr == '\'' || curr == '`' || curr == '^' || curr == '@' {
		return string(curr), nil
	}

	if curr == '"' {
		var builder strings.Builder
		builder.WriteByte('"')
		prevWasBackslash := false

		for {
			if t.Input == "" {
				return "", errors.New("unbalanced string quotes")
			}

			r, err := t.advanceChar()
			if err != nil {
				return "", err
			}

			if prevWasBackslash {
				if r == '"' {
					builder.WriteByte('"')
				} else if r == 'n' {
					builder.WriteByte('\n')
				} else if r == '\\' {
					builder.WriteByte('\\')
				} else {
					builder.WriteByte('\\')
					builder.WriteRune(r)
				}
				prevWasBackslash = false
			} else {
				if r == '"' {
					break
				} else if r == '\\' {
					prevWasBackslash = true
				} else {
					builder.WriteRune(r)
				}
			}
		}

		builder.WriteRune('"')
		return builder.String(), nil
	}

	if curr == ';' {
		t.Input = ""
		ret := t.Input
		return ret, nil
	}

	var builder strings.Builder
	builder.WriteRune(curr)
	for {
		if t.Input == "" {
			break
		}
		r, err := t.advanceChar()
		if err != nil {
			return "", err
		}

		if unicode.IsSpace(r) || r == ',' || r == '[' || r == ']' || r == '{' || r == '}' || r == '(' || r == ')' || r == '\'' || r == '`' || r == '^' || r == '@' || r == '~' || r == '"' || r == ';' {
			t.unadvanceChar(r)
			break
		}

		builder.WriteRune(r);
	}
	return builder.String(), nil
}

func ReadStr(input string) (MalType, error) {
	tokens := newTokenizer(input)
	return readForm(&tokens)
}

func readForm(tokens *tokenizer) (MalType, error) {
	first, err := tokens.peek()
	if err != nil {
		return nil, err
	}
	if first == "" {
		return nil, nil
	}

	if first == "(" || first == "[" {
		return readList(tokens)
	} else if first == "{" {
		return readMap(tokens)
	} else if first == "'" || first == "`" || first == "~" || first == "~@" || first == "@" {
		return readQuote(tokens)
	} else if first == "^" {
		return readMeta(tokens)
	} else {
		return readAtom(tokens)
	}
}

func readMeta(tokens *tokenizer) (MalType, error) {
	_, err := tokens.next()
	if err != nil {
		return nil, err
	}

	next1, err := readForm(tokens)
	if err != nil {
		return nil, err
	}
	next2, err := readForm(tokens)
	if err != nil {
		return nil, err
	}

	elements := []MalType{}
	elements = append(elements, MalTypeSymbol("with-meta"))
	elements = append(elements, next2)
	elements = append(elements, next1)
	return MalTypeList(elements), nil
}

func readQuote(tokens *tokenizer) (MalType, error) {
	first, err := tokens.next()
	if err != nil {
		return nil, err
	}

	var quote string
	switch (first) {
	case "'":
		quote = "quote"
	case "`":
		quote = "quasiquote"
	case "~":
		quote = "unquote"
	case "~@":
		quote = "splice-unquote"
	case "@":
		quote = "deref"
	default:
		panic("invalid token for quote")
	}

	next, err := readForm(tokens)
	if err != nil {
		return nil, err
	}

	elements := []MalType{}
	elements = append(elements, MalTypeSymbol(quote))
	elements = append(elements, next)
	return MalTypeList(elements), nil
}

func readMap(tokens *tokenizer) (MalType, error) {
	_, err := tokens.next()
	if err != nil {
		return nil, err
	}

	elements := make(map[MalType]MalType)
	for {
		str, err := tokens.peek()
		if err != nil {
			return nil, err
		}
		if str == "}" {
			break
		}
		if str == "" {
			return nil, errors.New("unbalanced parenthesis")
		}

		key, err := readForm(tokens)
		if err != nil {
			return nil, err
		}
		switch key.(type) {
		case MalTypeList, MalTypeVector, MalTypeHashMap:
			return nil, errors.New("invalid type for hashmap key")
		}

		value, err := readForm(tokens)
		if err != nil {
			return nil, err
		}

		elements[key] = value
	}

	_, err = tokens.next()
	if err != nil {
		return nil, err
	}

	return MalTypeHashMap(elements), nil
}

func readList(tokens *tokenizer) (MalType, error) {
	start, err := tokens.next()
	if err != nil {
		return nil, err
	}

	elements := []MalType{}
	for {
		str, err := tokens.peek()
		if err != nil {
			return nil, err
		}
		if (start == "(" && str == ")") || (start == "[" && str == "]") {
			break
		}
		if str == "" {
			return nil, errors.New("unbalanced parenthesis")
		}

		ret, err := readForm(tokens)
		if err != nil {
			return nil, err
		}
		elements = append(elements, ret)
	}

	_, err = tokens.next()
	if err != nil {
		return nil, err
	}

	if start == "[" {
		return MalTypeVector(elements), nil
	}
	return MalTypeList(elements), nil
}

func readAtom(tokens *tokenizer) (MalType, error) {
	token, err := tokens.next()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(token, "\"") {
		return MalTypeString(token[1:len(token)-1]), nil
	} else if strings.HasPrefix(token, ":") {
		return MalTypeKeyword(token[1:len(token)]), nil
	} else if token == "nil" {
		return MalTypeNil{}, nil
	} else if token == "true" {
		return MalTypeBoolean(true), nil
	} else if token == "false" {
		return MalTypeBoolean(false), nil
	}

	{
		val, err := strconv.ParseInt(token, 10, 64)
		if err == nil {
			return MalTypeInteger(val), nil
		}
	}

	return MalTypeSymbol(token), nil
}
