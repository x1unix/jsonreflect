package jsonx

import (
	"unicode"
)

var (
	nullVal  = []byte("null")
	trueVal  = []byte("true")
	falseVal = []byte("false")
)

type token = byte

const (
	tokenString       token = '"'
	tokenKeyDelimiter token = ':'
	tokenDelimiter    token = ','
	tokenObjectStart  token = '{'
	tokenObjectClose  token = '}'
	tokenArrayStart   token = '['
	tokenArrayClose   token = ']'
	tokenOther        token = 1
	tokenEnd          token = 0
)

const (
	charNumberNegative = '-'
)

type Parser struct {
	src []byte
	end int
}

func NewParser(src []byte) *Parser {
	return &Parser{
		src: src,
		end: len(src),
	}
}

func (p Parser) hasElem(idx int) bool {
	if len(p.src) <= idx {
		return false
	}
	return true
}

func (p *Parser) Parse() (Value, error) {
	v, err := p.parseValue(0, true)
	if err != nil {
		return nil, err
	}

	if v == nil {
		// skip empty document check
		return nil, nil
	}

	// throw error if something left after JSON contents
	pos := v.Ref()
	if p.end > pos.End {
		got, ok := p.getPosUntilNextNonDelimiter(pos.End + 1)
		if ok {
			return nil, NewInvalidExprError(got, p.end, p.src[got:])
		}
	}
	return v, nil
}

func (p Parser) getStartTokenAtPos(start int) (token, int, bool) {
	for i := start; i < p.end; i++ {
		switch t := p.src[i]; t {
		case '\t', '\r', '\n', ' ':
			// skip indentation
			continue
		case tokenString,
			tokenObjectStart,
			tokenArrayStart:
			return t, i, false
		default:
			return tokenOther, i, false
		}
	}
	return 0, start, true
}

func (p *Parser) parseValue(start int, root bool) (Value, error) {
	tkn, pos, end := p.getStartTokenAtPos(start)
	if end {
		// return nil for empty document
		return nil, nil
	}

	switch tkn {
	case tokenOther:
		return p.decodeScalarValue(pos, root)
	case tokenString:
		return p.decodeString(pos)
	case tokenArrayStart:
		return p.decodeArray(pos)
	case tokenObjectStart:
		return p.decodeObject(pos)
	default:
		return nil, NewUnexpectedCharacterError(start, pos, tkn)
	}
}

const (
	objectExpectKey = iota
	objectExpectDelimiter
	objectExpectValue
)

func (p Parser) decodeObject(start int) (*Object, error) {
	var lastKey string
	elems := make(map[string]Value, 0)
	curPos := start + 1 // next element should be after "{"
	expect := objectExpectKey

loop:
	for {
		if !p.hasElem(curPos) {
			return nil, NewParseError(newPosition(start, curPos), "unterminated object")
		}

		pos, ok := p.getPosUntilNextNonDelimiter(curPos)
		if !ok {
			return nil, NewParseError(newPosition(start, pos), "unterminated object")
		}

		char := p.src[pos]
		hadComma := false

		switch expect {
		case objectExpectDelimiter:
			if char != tokenKeyDelimiter {
				return nil, NewInvalidExprError(start, pos, []byte{char})
			}
			expect = objectExpectValue
			curPos++
		case objectExpectKey:
			switch char {
			case tokenObjectClose:
				if hadComma {
					// no trailing comma before object close
					return nil, NewUnexpectedCharacterError(start, pos, char)
				}
				break loop
			case tokenDelimiter:
				if len(elems) == 0 || hadComma {
					// no multiple commas after prop
					return nil, NewUnexpectedCharacterError(start, pos, char)
				}
				hadComma = true
				curPos++
			case tokenString:
				hadComma = false
				str, err := p.decodeString(pos)
				if err != nil {
					return nil, err
				}

				lastKey, err = str.String()
				if err != nil {
					return nil, NewParseError(newPosition(start, pos), err.Error())
				}

				curPos = str.Position.End + 1
				expect = objectExpectDelimiter
			default:
				return nil, NewUnexpectedCharacterError(start, pos, char)
			}
		case objectExpectValue:
			val, err := p.parseValue(pos, false)
			if err != nil {
				return nil, err
			}
			curPos = val.Ref().End + 1
			elems[lastKey] = val
			expect = objectExpectKey
		}
	}

	return newObject(start, curPos, elems), nil
}

func (p Parser) decodeArray(start int) (*Array, error) {
	var elems []Value
	curPos := start + 1      // next element should be after "[" char
	prevIsDelimiter := false // handle trailing commas
	for {
		if !p.hasElem(curPos) {
			return nil, NewParseError(newPosition(start, curPos), "unterminated array statement")
		}

		switch char := p.src[curPos]; char {
		case '\t', '\r', '\n', ' ':
			curPos++
			continue
		case tokenDelimiter:
			if prevIsDelimiter {
				return nil, NewUnexpectedCharacterError(curPos-1, curPos, tokenDelimiter)
			}

			prevIsDelimiter = true
			curPos++
		case tokenArrayClose:
			if prevIsDelimiter {
				return nil, NewUnexpectedCharacterError(curPos-1, curPos, tokenDelimiter)
			}
			return newArray(newPosition(start, curPos), elems...), nil
		default:
			prevIsDelimiter = false
			val, err := p.parseValue(curPos, false)
			if err != nil {
				return nil, err
			}

			if elems == nil {
				// allocate slice of values only if necessary
				elems = make([]Value, 0, 2)
			}

			curPos = val.Ref().End + 1
			elems = append(elems, val)
		}
	}
}

func (p Parser) decodeString(start int) (*String, error) {
	end := start
	hasEscape := false
	complete := false
outer:
	for i := start + 1; i < p.end; i++ {
		char := p.src[i]
		switch char {
		case tokenString:
			if !hasEscape {
				end = i
				complete = true
				break outer
			}

			continue
		case '\\':
			if hasEscape {
				// double escape
				hasEscape = false
				continue
			}
			hasEscape = true
		default:
			if hasEscape {
				hasEscape = false
			}
			continue
		}
	}

	if !complete {
		endPos := p.getPosUntilNextDelimiter(start)
		return nil, NewParseError(newPosition(start, endPos), "unterminated string '%s'", p.src[start:endPos])
	}

	return newString(newPosition(start, end), p.src[start:end+1]), nil
}

func (p Parser) decodeNumber(start int) (*Number, error) {
	// keep track of chars that should appear only once
	once := make(map[byte]struct{}, 2)

	var end int
outer:
	for i := start; i < p.end; i++ {
		char := p.src[i]
		switch char {
		case '\t', '\r', '\n', ' ', ',', tokenObjectClose, tokenArrayClose:
			break outer
		case '.', charNumberNegative:
			// chars '-' and '.' should appear once in numbers
			if _, ok := once[char]; ok {
				endPos := p.getPosUntilNextDelimiter(start)
				return nil, NewInvalidExprError(start, endPos, p.src[start:endPos])
			}

			once[char] = struct{}{}
			end = i
		default:
			if unicode.IsNumber(rune(char)) {
				end = i
				continue
			}

			endPos := p.getPosUntilNextDelimiter(start)
			return nil, NewInvalidExprError(start, endPos, p.src[start:endPos])
		}
	}

	str := p.src[start : end+1]
	pos := Position{
		Start: start,
		End:   end,
	}
	return ParseNumber(pos, string(str), 64)
}

func (p Parser) decodeScalarValue(start int, root bool) (Value, error) {
	// numbers can start with number (obviously) or negative symbol (-)
	if char := p.src[start]; unicode.IsNumber(rune(char)) || char == charNumberNegative {
		return p.decodeNumber(start)
	}

	// other possible scalar values are: false, true and null
	var (
		match          []byte = nil
		possibleResult Value
	)

	char := p.src[start]
	exprEnd := p.getPosUntilNextDelimiter(start)
	switch char {
	case trueVal[0]:
		match = trueVal
		possibleResult = newBoolean(newPosition(start, start+len(trueVal)-1), true)
	case falseVal[0]:
		match = falseVal
		possibleResult = newBoolean(newPosition(start, start+len(falseVal)-1), false)
	case nullVal[0]:
		match = nullVal
		possibleResult = newNull(newPosition(start, start+len(nullVal)-1))
	default:
		return nil, NewUnexpectedCharacterError(start, start+1, char)
	}

	if root {
		// expression might start correctly but contain invalid values like:
		// "nullsomething" or "fals"
		expectEnd := start + len(match)
		if expectEnd != exprEnd {
			return nil, NewInvalidExprError(start, exprEnd, p.src[start:exprEnd])
		}
	}

	str := p.src[start:exprEnd]
	if string(str) != string(match) {
		return nil, NewInvalidExprError(start, exprEnd, p.src[start:exprEnd])
	}

	return possibleResult, nil
}

func (p Parser) getPosUntilNextNonDelimiter(start int) (int, bool) {
	for i := start; i < p.end; i++ {
		switch p.src[i] {
		case '\t', '\r', '\n', ' ':
			continue
		default:
			return i, true
		}
	}
	return 0, false
}

func (p Parser) getPosUntilNextDelimiter(start int) int {
	lastChar := start
	for i := start; i < p.end; i++ {
		switch p.src[i] {
		case '\t', '\r', '\n', ' ', tokenDelimiter, tokenArrayClose, tokenObjectClose:
			//if i == start {
			//	return start
			//}
			return i
		default:
			lastChar = i + 1
			continue
		}
	}
	return lastChar
}

func (p Parser) isCharDelimiterOrPadding(index int) bool {
	if p.end <= index {
		return true
	}

	switch p.src[index] {
	case '\t', '\r', '\n', ' ', tokenDelimiter:
		return true
	default:
		return false
	}
}
