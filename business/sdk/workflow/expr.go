package workflow

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// EvalExpr evaluates a simple arithmetic expression with variables from vars.
// Supports: +, -, *, /, %, parentheses, float/integer literals, and variable names.
// Variable names resolve from vars; values must be numeric or string-numeric.
// On any error (unknown variable, division by zero, parse error), returns (0, err).
func EvalExpr(expr string, vars map[string]any) (float64, error) {
	p := &exprParser{input: strings.TrimSpace(expr), vars: vars}
	result, err := p.parseAddSub()
	if err != nil {
		return 0, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return 0, fmt.Errorf("unexpected character %q at position %d", p.input[p.pos], p.pos)
	}
	return result, nil
}

type exprParser struct {
	input string
	pos   int
	vars  map[string]any
}

func (p *exprParser) parseAddSub() (float64, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return 0, err
	}
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		op := p.input[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		right, err := p.parseMulDiv()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *exprParser) parseMulDiv() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		op := p.input[p.pos]
		if op != '*' && op != '/' && op != '%' {
			break
		}
		p.pos++
		right, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		switch op {
		case '*':
			left *= right
		case '/':
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		case '%':
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		}
	}
	return left, nil
}

func (p *exprParser) parseUnary() (float64, error) {
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '-' {
		p.pos++
		val, err := p.parsePrimary()
		return -val, err
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (float64, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return 0, fmt.Errorf("unexpected end of expression")
	}
	ch := p.input[p.pos]

	// Parenthesized expression.
	if ch == '(' {
		p.pos++
		val, err := p.parseAddSub()
		if err != nil {
			return 0, err
		}
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return val, nil
	}

	// Number literal.
	if unicode.IsDigit(rune(ch)) || ch == '.' {
		start := p.pos
		for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
			p.pos++
		}
		return toFloat64(p.input[start:p.pos])
	}

	// Variable name (letters, digits, underscores; must start with letter or underscore).
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		start := p.pos
		for p.pos < len(p.input) && (unicode.IsLetter(rune(p.input[p.pos])) || unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '_') {
			p.pos++
		}
		name := p.input[start:p.pos]
		val, ok := p.vars[name]
		if !ok {
			return 0, fmt.Errorf("unknown variable %q", name)
		}
		return toFloat64(val)
	}

	return 0, fmt.Errorf("unexpected character %q at position %d", ch, p.pos)
}

func (p *exprParser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}
