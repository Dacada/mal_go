package common

import (
	"fmt"
	"strconv"
	"strings"
)

func PrStr(input MalType, print_readably bool) string {
	if input == nil {
		return ""
	}

	switch i := input.(type) {
	case MalTypeList, MalTypeVector:
		return prList(input, print_readably)
	case MalTypeHashMap:
		return prMap(i, print_readably)
	case MalTypeInteger:
		return strconv.FormatInt(int64(i), 10)
	case MalTypeNil:
		return "nil"
	case MalTypeBoolean:
		if i {
			return "true"
		} else {
			return "false"
		}
	case MalTypeKeyword:
		return ":" + string(i)
	case MalTypeString:
		return prStr(i, print_readably)
	case MalTypeSymbol:
		return string(i)
	case MalTypeFunction:
		return "#<function>"
	case MalTypeTCOFunction:
		return "#<function>"
	case MalTypeAtom:
		return prAtom(i, print_readably)
	default:
		panic(fmt.Sprintf("invalid mal type %T", input))
	}
}

func prAtom(input MalTypeAtom, print_readably bool) string {
	var l MalTypeList
	l = append(l, MalTypeSymbol("atom"))
	l = append(l, *input)
	return prList(l, print_readably)
}

func prStr(input MalTypeString, print_readably bool) string {
	var builder strings.Builder
	if print_readably {
		builder.WriteByte('"')
	}

	for _, char := range input {
		if print_readably {
			if char == '"' {
				builder.WriteByte('\\')
				builder.WriteByte('"')
			} else if char == '\\' {
				builder.WriteByte('\\')
				builder.WriteByte('\\')
			} else if char == '\n' {
				builder.WriteByte('\\')
				builder.WriteByte('n')
			} else {
				builder.WriteRune(char)
			}
		} else {
			builder.WriteRune(char)
		}
	}

	if print_readably {
		builder.WriteByte('"')
	}
	return builder.String()
}

func prMap(input MalTypeHashMap, print_readably bool) string {
	var builder strings.Builder
	builder.WriteByte('{')
	for k := range input {
		builder.WriteString(PrStr(k, print_readably))
		builder.WriteByte(' ')
		builder.WriteString(PrStr(input[k], print_readably))
		builder.WriteByte(' ')
	}
	if len(input) > 0 {
		buf := builder.String()
		buf = buf[:len(buf)-1]
		builder.Reset()
		builder.WriteString(buf)
	}
	builder.WriteByte('}')
	return builder.String()
}

func prList(input MalType, print_readably bool) string {
	var start rune
	var end rune
	var list []MalType
	switch input.(type) {
	case MalTypeList:
		start = '('
		end = ')'
		list = []MalType(input.(MalTypeList))
	case MalTypeVector:
		start = '['
		end = ']'
		list = []MalType(input.(MalTypeVector))
	default:
		panic("invalid list type")
	}
	
	var builder strings.Builder
	builder.WriteRune(start)
	
	if len(list) > 0 {
		builder.WriteString(PrStr(list[0], print_readably))
	}
	for i := 1; i < len(list); i++ {
		builder.WriteByte(' ')
		builder.WriteString(PrStr(list[i], print_readably))
	}
	builder.WriteRune(end)
	return builder.String()
}
