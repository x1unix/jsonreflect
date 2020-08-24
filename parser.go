package jsonx

import (
	"fmt"
	"unicode"
)

var (
	nullVal  = []byte("null")
	trueVal  = []byte("true")
	falseVal = []byte("false")
)

type token = byte

const (
	tokenString      token = '"'
	tokenValue       token = ':'
	tokenDelimiter   token = ','
	tokenObjectStart token = '{'
	tokenObjectClose token = '}'
	tokenArrayStart  token = '['
	tokenArrayClose  token = ']'
	tokenOther       token = 0
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
	return nil, fmt.Errorf("not implemented")
}

func (p Parser) getStartTokenAtPos(start int) (token, int, error) {
	for i := start; i < p.end; i++ {
		switch t := p.src[i]; t {
		case '\t', '\r', '\n', ' ':
			// skip indentation
			continue
		case tokenString,
			tokenObjectStart,
			tokenArrayStart:
			return t, i, nil
		default:
			return tokenOther, i, nil
		}
	}
	return tokenOther, start, nil
}

func (p *Parser) parseValue(start int) (Value, error) {
	tkn, pos, err := p.getStartTokenAtPos(start)
	if err != nil {
		return nil, err
	}

	switch tkn {
	case tokenOther:
		return p.decodeScalarValue(pos)
	case tokenString:
		return p.decodeString(pos)
	case tokenArrayStart:
		return p.decodeArray(pos)
	default:
		return nil, NewUnexpectedCharacterError(start, pos, tkn)
	}
}

func (p Parser) decodeArray(start int) (*Array, error) {
	elems := make([]Value, 0, 2)
	curPos := start + 1 // next element should be after "[" char
	for {
		if p.hasElem(curPos) {
			return nil, NewParseError(newPosition(start, start), "unterminated array statement")
		}

		switch char := p.src[curPos]; char {
		case tokenDelimiter:
			if !p.hasElem(curPos + 1) {
				return nil, NewParseError(newPosition(start, curPos), "unterminated array statement")
			}
			val, err := p.parseValue(curPos + 1)
			if err != nil {
				return nil, err
			}

			curPos = val.Ref().End + 1
			elems = append(elems, val)
		case tokenArrayClose:
			return newArray(newPosition(start, curPos), elems), nil
		default:
			return nil, NewUnexpectedCharacterError(start, curPos, char)
		}
	}
}

func (p Parser) decodeString(start int) (*String, error) {
	end := 0
	hasEscape := false
outer:
	for i := 1; i <= p.end; i++ {
		char := p.src[i]
		switch char {
		case tokenString:
			if !hasEscape {
				end = start + i
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

	return newString(newPosition(start, end), p.src[start:end]), nil
}

func (p Parser) decodeNumber(start int) (*Number, error) {
	var end int
outer:
	for i := start; i <= p.end; i++ {
		char := p.src[i]
		switch char {
		case '\t', '\r', '\n', ' ', ',':
			break outer
		case '.':
			end = i
		default:
			if unicode.IsNumber(rune(char)) {
				end = i
				continue
			}

			return nil, NewUnexpectedCharacterError(start, i, char)
		}
	}

	str := p.src[start:end]
	pos := Position{
		Start: start,
		End:   end,
	}
	return ParseNumber(pos, string(str), 64)
}

func (p Parser) decodeScalarValue(start int) (Value, error) {
	if unicode.IsNumber(rune(p.src[start])) {
		return p.decodeNumber(start)
	}

	// other possible scalar values are: false, true and null
	var (
		match          []byte = nil
		possibleResult Value
		char           = p.src[start]
	)
	switch char {
	case trueVal[0]:
		match = trueVal
		possibleResult = newBoolean(newPosition(start, start+len(falseVal)), true)
	case falseVal[0]:
		match = falseVal
		possibleResult = newBoolean(newPosition(start, start+len(falseVal)), false)
	case nullVal[0]:
		match = nullVal
		possibleResult = newNull(newPosition(start, start+len(falseVal)))
	default:
		return nil, NewUnexpectedCharacterError(start, start+1, char)
	}

	end := start + len(match)
	if len(p.src) < end {
		return nil, NewUnexpectedCharacterError(start, len(p.src), char)
	}

	str := p.src[start:end]
	if string(str) != string(nullVal) {
		return nil, NewUnexpectedCharacterError(start, end, char)
	}

	return possibleResult, nil
}

func isNullToken(src []byte, start int) bool {
	end := start + len(nullVal)
	if len(src) < end {
		return false
	}

	str := src[start:end]
	return string(str) == string(nullVal)
}
