// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/errs"
)

var legacyEvalEmbeddedRegex = regexp.MustCompile(`\|\|[^|]+\|\|`)

type operator struct {
	symbol     string
	call       string // When set, the operator is emitted as a call to this function rather than as an infix operator.
	precedence int
	unary      bool
}

type parsedFunction struct {
	unaryOp  *operator
	function string
	args     string
}

type expressionOperand struct {
	unaryOp *operator
	value   string
}

type expressionOperator struct {
	op      *operator
	unaryOp *operator
}

type expressionTree struct {
	left      any
	right     any
	op        *operator
	unaryOp   *operator
	hasParens bool
}

type exprToScript struct {
	operators     []*operator
	functions     map[string]string
	operandStack  []any
	operatorStack []*expressionOperator
}

// EmbeddedExprToScript converts an old-style embedded expression string into an embedded JavaScript script string.
func EmbeddedExprToScript(text string) string {
	return legacyEvalEmbeddedRegex.ReplaceAllStringFunc(text, func(s string) string {
		// Embedded scripts now use <script>...</script> to avoid conflicts with the OR operator: ||
		return scriptStart + ExprToScript(s[2:len(s)-2]) + scriptEnd
	})
}

// ExprToScript converts an old-style expression string into a JavaScript script string.
func ExprToScript(expr string) string {
	e := exprToScript{
		operators: []*operator{
			{symbol: "("},
			{symbol: ")"},
			{symbol: "||", precedence: 1},
			{symbol: "&&", precedence: 2},
			{symbol: "!=", precedence: 3},
			{symbol: "!", unary: true},
			{symbol: "==", precedence: 3},
			{symbol: ">=", precedence: 4},
			{symbol: ">", precedence: 4},
			{symbol: "<=", precedence: 4},
			{symbol: "<", precedence: 4},
			{symbol: "+", precedence: 5, unary: true},
			{symbol: "-", precedence: 5, unary: true},
			{symbol: "*", precedence: 6},
			{symbol: "/", precedence: 6},
			{symbol: "%", precedence: 6},
			// JavaScript's ** is not a usable stand-in for the legacy exponent operator: it is right-associative
			// where the legacy operator was left-associative (2^3^2 is 64, not 512), and its grammar rejects an
			// unparenthesized unary operand on its left, making "-$st ** 2" a syntax error. Math.pow has neither
			// problem, since the tree already carries the intended grouping.
			{symbol: "^", call: "Math.pow", precedence: 7},
		},
		functions: map[string]string{
			"if":     "iff",
			"signed": "signedValue",

			"advantage_level": "entity.traitLevel",
			"enc":             "entity.currentEncumbrance",
			"has_trait":       "entity.hasTrait",
			"random_height":   "entity.randomHeightInInches",
			"random_weight":   "entity.randomWeightInPounds",
			"skill_level":     "entity.skillLevel",
			"trait_level":     "entity.traitLevel",
			"weapon_damage":   "entity.weaponDamage",

			"add_dice":        "dice.add",
			"dice_count":      "dice.count",
			"dice_modifier":   "dice.modifier",
			"dice_multiplier": "dice.multiplier",
			"dice_sides":      "dice.sides",
			"dice":            "dice.from",
			"roll":            "dice.roll",
			"subtract_dice":   "dice.subtract",

			"abs":   "Math.abs",
			"cbrt":  "Math.cbrt",
			"ceil":  "Math.ceil",
			"exp":   "Math.exp",
			"exp2":  "Math.exp2",
			"floor": "Math.floor",
			"log":   "Math.log",
			"log10": "Math.log10",
			"log1p": "Math.log1p",
			"max":   "Math.max",
			"min":   "Math.min",
			"round": "Math.round",
			"sqrt":  "Math.sqrt",

			"ssrt_to_yards": "measure.modifierToYards",
			"ssrt":          "measure.modifier",
		},
	}
	// If the expression can't be fully converted, leave the original text alone.
	if parts, err := e.processExpression(nil, expr); err == nil && len(parts) != 0 {
		expr = strings.Join(parts, " ")
	}
	return expr
}

func (e *exprToScript) processExpression(parts []string, expression string) ([]string, error) {
	if err := e.parse(expression); err != nil {
		return parts, err
	}
	for len(e.operatorStack) != 0 {
		e.processTree()
	}
	if len(e.operandStack) == 0 {
		return parts, nil
	}
	for _, op := range e.operandStack {
		var err error
		if parts, err = e.process(parts, op); err != nil {
			return parts, err
		}
	}
	return parts, nil
}

func (e *exprToScript) parse(expression string) error {
	var unaryOp *operator
	haveOperand := false
	e.operandStack = nil
	e.operatorStack = nil
	i := 0
	for i < len(expression) {
		ch := expression[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}
		opIndex, op := e.nextOperator(expression, i, nil)
		if opIndex > i || opIndex == -1 {
			var err error
			if i, err = e.processOperand(expression, i, opIndex, unaryOp); err != nil {
				return err
			}
			haveOperand = true
			unaryOp = nil
		}
		if opIndex == i {
			if op != nil && op.unary && !haveOperand {
				i = opIndex + len(op.symbol)
				if unaryOp != nil {
					return errs.Newf("consecutive unary operators are not allowed at index %d", i)
				}
				unaryOp = op
			} else {
				// processOperator may consume more than the operator it was handed (a function call and the operator
				// that follows it), so the operator it reports back is the one that determines whether an operand is
				// now available.
				var err error
				var last *operator
				if i, last, err = e.processOperator(expression, opIndex, op, haveOperand, unaryOp); err != nil {
					return err
				}
				unaryOp = nil
				if last == nil || last.symbol != ")" {
					haveOperand = false
				}
			}
		}
	}
	return nil
}

func (e *exprToScript) nextOperator(expression string, start int, match *operator) (int, *operator) {
	for i := start; i < len(expression); i++ {
		if match != nil {
			if match.match(expression, i, len(expression)) {
				return i, match
			}
		} else {
			for _, op := range e.operators {
				if op.match(expression, i, len(expression)) {
					return i, op
				}
			}
		}
	}
	return -1, nil
}

func (e *exprToScript) processOperand(expression string, start, opIndex int, unaryOp *operator) (int, error) {
	if opIndex == -1 {
		text := strings.TrimSpace(expression[start:])
		if text == "" {
			return -1, errs.Newf("expression is invalid at index %d", start)
		}
		e.operandStack = append(e.operandStack, &expressionOperand{
			unaryOp: unaryOp,
			value:   text,
		})
		return len(expression), nil
	}
	text := strings.TrimSpace(expression[start:opIndex])
	if text == "" {
		return -1, errs.Newf("expression is invalid at index %d", start)
	}
	e.operandStack = append(e.operandStack, &expressionOperand{
		unaryOp: unaryOp,
		value:   text,
	})
	return opIndex, nil
}

// processOperator processes the operator at the given index, returning the index just past what was consumed along with
// the last operator that was actually processed, which is not necessarily the one that was passed in.
func (e *exprToScript) processOperator(expression string, index int, op *operator, haveOperand bool, unaryOp *operator) (int, *operator, error) {
	if haveOperand && op != nil && op.symbol == "(" {
		var err error
		var closeOp *operator
		index, closeOp, err = e.processFunction(expression, index)
		if err != nil {
			return -1, nil, err
		}
		index += len(closeOp.symbol)
		next, nextOp := e.nextOperator(expression, index, nil)
		if nextOp == nil {
			return index, closeOp, nil
		}
		op = nextOp
		index = next
	}
	switch op.symbol {
	case "(":
		e.operatorStack = append(e.operatorStack, &expressionOperator{
			op:      op,
			unaryOp: unaryOp,
		})
	case ")":
		var stackOp *expressionOperator
		if len(e.operatorStack) > 0 {
			stackOp = e.operatorStack[len(e.operatorStack)-1]
		}
		for stackOp != nil && stackOp.op.symbol != "(" {
			e.processTree()
			if len(e.operatorStack) > 0 {
				stackOp = e.operatorStack[len(e.operatorStack)-1]
			} else {
				stackOp = nil
			}
		}
		if len(e.operatorStack) == 0 {
			return -1, nil, errs.Newf("invalid expression at index %d", index)
		}
		stackOp = e.operatorStack[len(e.operatorStack)-1]
		if stackOp.op.symbol != "(" {
			return -1, nil, errs.Newf("invalid expression at index %d", index)
		}
		e.operatorStack = e.operatorStack[:len(e.operatorStack)-1]
		if stackOp.unaryOp != nil {
			if len(e.operandStack) == 0 {
				return -1, nil, errs.Newf("invalid expression at index %d", index)
			}
			left := e.operandStack[len(e.operandStack)-1]
			e.operandStack = e.operandStack[:len(e.operandStack)-1]
			e.operandStack = append(e.operandStack, &expressionTree{
				left:    left,
				unaryOp: stackOp.unaryOp,
			})
		}
	default:
		if len(e.operatorStack) > 0 {
			stackOp := e.operatorStack[len(e.operatorStack)-1]
			for stackOp != nil && stackOp.op.precedence >= op.precedence {
				e.processTree()
				if len(e.operatorStack) > 0 {
					stackOp = e.operatorStack[len(e.operatorStack)-1]
				} else {
					stackOp = nil
				}
			}
		}
		e.operatorStack = append(e.operatorStack, &expressionOperator{
			op:      op,
			unaryOp: unaryOp,
		})
	}
	return index + len(op.symbol), op, nil
}

func (e *exprToScript) processFunction(expression string, opIndex int) (int, *operator, error) {
	var op *operator
	parens := 1
	next := opIndex
	for parens > 0 {
		if next, op = e.nextOperator(expression, next+1, nil); op == nil {
			return -1, nil, errs.Newf("function not closed at index %d", opIndex)
		}
		switch op.symbol {
		case "(":
			parens++
		case ")":
			parens--
		default:
		}
	}
	if len(e.operandStack) == 0 {
		return -1, nil, errs.Newf("invalid stack at index %d", next)
	}
	operand, ok := e.operandStack[len(e.operandStack)-1].(*expressionOperand)
	if !ok {
		return -1, nil, errs.Newf("unexpected operand stack value at index %d", next)
	}
	e.operandStack = e.operandStack[:len(e.operandStack)-1]
	f, exists := e.functions[operand.value]
	if !exists {
		return -1, nil, errs.Newf("function not defined: %s", operand.value)
	}
	e.operandStack = append(e.operandStack, &parsedFunction{
		function: f,
		args:     expression[opIndex+1 : next],
		unaryOp:  operand.unaryOp,
	})
	return next, op, nil
}

func (e *exprToScript) processTree() {
	var right any
	if len(e.operandStack) > 0 {
		right = e.operandStack[len(e.operandStack)-1]
		e.operandStack = e.operandStack[:len(e.operandStack)-1]
	}
	var left any
	if len(e.operandStack) > 0 {
		left = e.operandStack[len(e.operandStack)-1]
		e.operandStack = e.operandStack[:len(e.operandStack)-1]
	}
	op := e.operatorStack[len(e.operatorStack)-1]
	e.operatorStack = e.operatorStack[:len(e.operatorStack)-1]
	e.operandStack = append(e.operandStack, &expressionTree{
		left:      left,
		right:     right,
		op:        op.op,
		hasParens: len(e.operatorStack) > 0 && e.operatorStack[len(e.operatorStack)-1].op.symbol == "(",
	})
}

func (e *exprToScript) process(parts []string, op any) ([]string, error) {
	switch v := op.(type) {
	case *expressionTree:
		start := len(parts)
		var err error
		if v.left != nil && v.right != nil {
			if v.op != nil && v.op.call != "" {
				if parts, err = e.processCall(parts, v); err != nil {
					return parts, err
				}
			} else {
				if parts, err = e.process(parts, v.left); err != nil {
					return parts, err
				}
				if v.op != nil {
					parts = append(parts, v.op.String())
				}
				if parts, err = e.process(parts, v.right); err != nil {
					return parts, err
				}
			}
		} else {
			unaryOp := v.unaryOp
			if unaryOp == nil && v.op != nil && v.op.unary {
				unaryOp = v.op
			}
			operand := v.left
			if operand == nil {
				operand = v.right
			}
			if unaryOp == nil || operand == nil {
				return parts, errs.New("expression is invalid")
			}
			if parts, err = e.process(parts, operand); err != nil {
				return parts, err
			}
			if len(parts) == start {
				return parts, errs.New("expression is invalid")
			}
			parts[start] = unaryOp.String() + parts[start]
		}
		if v.hasParens {
			if len(parts) == start {
				return parts, errs.New("expression is invalid")
			}
			parts[start] = "(" + parts[start]
			parts[len(parts)-1] = parts[len(parts)-1] + ")"
		}
		return parts, nil
	case *expressionOperand:
		value := v.value
		if value != "true" && value != "false" {
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				if !strings.HasPrefix(value, "$") && (!strings.HasPrefix(value, `"`) || !strings.HasSuffix(value, `"`)) {
					value = `"` + value + `"`
				}
			}
		}
		if v.unaryOp != nil {
			value = v.unaryOp.String() + value
		}
		return append(parts, value), nil
	case *parsedFunction:
		var funcParts []string
		i := 0
		skip := false
		next, remaining := extractNextEvalArg(v.args)
		for next != "" {
			// Special case for some dice functions
			if strings.HasPrefix(v.function, "dice.") && v.function != "dice.from" {
				if i == 0 || i > 0 && v.function != "dice.roll" {
					funcParts, skip = extractDiceArg(funcParts, next)
				}
			} else if strings.HasPrefix(next, `"`) && strings.HasSuffix(next, `"`) {
				// If the next argument is a string, we don't need to do anything special
				funcParts = append(funcParts, next)
				skip = true
			}
			i++
			if skip {
				skip = false
			} else {
				mark := len(funcParts)
				var err error
				if funcParts, err = e.processExpression(funcParts, next); err != nil {
					return parts, err
				}
				if len(funcParts) == mark {
					return parts, errs.Newf("argument is invalid: %s", next)
				}
			}
			next, remaining = extractNextEvalArg(remaining)
			if next != "" && len(funcParts) != 0 {
				funcParts[len(funcParts)-1] = funcParts[len(funcParts)-1] + ","
			}
		}
		call := v.function + "(" + strings.Join(funcParts, " ") + ")"
		if v.unaryOp != nil {
			call = v.unaryOp.String() + call
		}
		return append(parts, call), nil
	default:
		if op != nil {
			return append(parts, "<unknown>"), nil
		}
		return parts, nil
	}
}

// processCall emits a binary operator as a call to the function named by the operator, with the left and right sides as
// its two arguments. The tree already encodes the grouping the legacy expression called for, so the call form carries
// it over without depending on the precedence or associativity of any JavaScript operator.
func (e *exprToScript) processCall(parts []string, v *expressionTree) ([]string, error) {
	start := len(parts)
	var err error
	if parts, err = e.process(parts, v.left); err != nil {
		return parts, err
	}
	mid := len(parts)
	if mid == start {
		return parts, errs.New("expression is invalid")
	}
	if parts, err = e.process(parts, v.right); err != nil {
		return parts, err
	}
	if len(parts) == mid {
		return parts, errs.New("expression is invalid")
	}
	parts[start] = v.op.call + "(" + parts[start]
	parts[mid-1] += ","
	parts[len(parts)-1] += ")"
	return parts, nil
}

func extractDiceArg(parts []string, arg string) ([]string, bool) {
	if strings.IndexByte(arg, '(') != -1 {
		return parts, false
	}
	if !strings.HasPrefix(arg, "$") && (!strings.HasPrefix(arg, `"`) || !strings.HasSuffix(arg, `"`)) {
		arg = `"` + arg + `"`
	}
	return append(parts, arg), true
}

func extractNextEvalArg(args string) (arg, remaining string) {
	parens := 0
	for i, ch := range args {
		switch {
		case ch == '(':
			parens++
		case ch == ')':
			parens--
		case ch == ',' && parens == 0:
			return args[:i], args[i+1:]
		default:
		}
	}
	return args, ""
}

func (o *operator) String() string {
	return o.symbol
}

func (o *operator) match(expression string, start, maximum int) bool {
	if maximum-start < len(o.symbol) {
		return false
	}
	matches := o.symbol == expression[start:start+len(o.symbol)]
	// Hack to allow negative exponents on floating point numbers (i.e. 1.2e-2)
	if matches && len(o.symbol) == 1 && o.symbol == "-" && start > 1 && expression[start-1:start] == "e" {
		ch, _ := utf8.DecodeRuneInString(expression[start-2 : start-1])
		if unicode.IsDigit(ch) {
			return false
		}
	}
	return matches
}
