package searchquery

import (
	"time"
	"github.com/kamichidu/go-gae-search-query/ast"
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleQuery
	ruleExprs
	ruleExpr
	ruleProperty
	ruleOperator
	ruleValue
	ruleTime
	ruleString
	ruleBareString
	ruleQuotedString
	ruleInteger
	ruleFloat
	ruleBool
	ruleSpacing
	ruleComment
	ruleSpace
	ruleEndOfLine
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	rulePegText
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"Query",
	"Exprs",
	"Expr",
	"Property",
	"Operator",
	"Value",
	"Time",
	"String",
	"BareString",
	"QuotedString",
	"Integer",
	"Float",
	"Bool",
	"Spacing",
	"Comment",
	"Space",
	"EndOfLine",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"PegText",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type Query struct {
	astBuilder

	Buffer string
	buffer []rune
	rules  [45]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Query
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Query) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *Query) Highlighter() {
	p.PrintSyntax()
}

func (p *Query) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.reduceAnd()
		case ruleAction1:
			p.finalize()
		case ruleAction2:
			p.pushOr()
		case ruleAction3:
			p.pushOperatorExpr()
		case ruleAction4:
			p.pushColonExpr()
		case ruleAction5:
			p.pushNewState()
		case ruleAction6:
			p.reduceAnd()
		case ruleAction7:
			p.popNewState()
		case ruleAction8:
			p.pushNot()
		case ruleAction9:
			p.pushKeywordExpr()
		case ruleAction10:
			p.pushProperty(buffer[begin:end])
		case ruleAction11:
			p.pushOperator(ast.OpEq)
		case ruleAction12:
			p.pushOperator(ast.OpNeq)
		case ruleAction13:
			p.pushOperator(ast.OpNeq)
		case ruleAction14:
			p.pushOperator(ast.OpLt)
		case ruleAction15:
			p.pushOperator(ast.OpLe)
		case ruleAction16:
			p.pushOperator(ast.OpGt)
		case ruleAction17:
			p.pushOperator(ast.OpGe)
		case ruleAction18:
			p.pushTimeValue(time.RFC3339, buffer[begin:end])
		case ruleAction19:
			p.pushTimeValue("2006-01-02", buffer[begin:end])
		case ruleAction20:
			p.pushStringValue(buffer[begin:end])
		case ruleAction21:
			p.pushStringValue(buffer[begin:end])
		case ruleAction22:
			p.pushIntegerValue(buffer[begin:end])
		case ruleAction23:
			p.pushFloatValue(buffer[begin:end])
		case ruleAction24:
			p.pushBoolValue(true)
		case ruleAction25:
			p.pushBoolValue(false)

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Query) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Query <- <(Spacing Exprs Action0 Spacing !. Action1)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if !_rules[ruleExprs]() {
					goto l0
				}
				if !_rules[ruleAction0]() {
					goto l0
				}
				if !_rules[ruleSpacing]() {
					goto l0
				}
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
				if !_rules[ruleAction1]() {
					goto l0
				}
				depth--
				add(ruleQuery, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Exprs <- <(Expr ((Spacing ('A' 'N' 'D') Spacing Expr) / (Spacing ('O' 'R') Spacing Expr Action2) / (Spacing Expr))*)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !_rules[ruleExpr]() {
					goto l3
				}
			l5:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					{
						position7, tokenIndex7, depth7 := position, tokenIndex, depth
						if !_rules[ruleSpacing]() {
							goto l8
						}
						if buffer[position] != rune('A') {
							goto l8
						}
						position++
						if buffer[position] != rune('N') {
							goto l8
						}
						position++
						if buffer[position] != rune('D') {
							goto l8
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l8
						}
						if !_rules[ruleExpr]() {
							goto l8
						}
						goto l7
					l8:
						position, tokenIndex, depth = position7, tokenIndex7, depth7
						if !_rules[ruleSpacing]() {
							goto l9
						}
						if buffer[position] != rune('O') {
							goto l9
						}
						position++
						if buffer[position] != rune('R') {
							goto l9
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l9
						}
						if !_rules[ruleExpr]() {
							goto l9
						}
						if !_rules[ruleAction2]() {
							goto l9
						}
						goto l7
					l9:
						position, tokenIndex, depth = position7, tokenIndex7, depth7
						if !_rules[ruleSpacing]() {
							goto l6
						}
						if !_rules[ruleExpr]() {
							goto l6
						}
					}
				l7:
					goto l5
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				depth--
				add(ruleExprs, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 Expr <- <((Property Spacing ((Operator Spacing Value Action3) / (':' Spacing Expr Action4))) / ('(' Action5 Spacing Exprs Action6 Spacing ')' Action7) / ('N' 'O' 'T' Spacing Expr Action8) / (Value Action9))> */
		func() bool {
			position10, tokenIndex10, depth10 := position, tokenIndex, depth
			{
				position11 := position
				depth++
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if !_rules[ruleProperty]() {
						goto l13
					}
					if !_rules[ruleSpacing]() {
						goto l13
					}
					{
						position14, tokenIndex14, depth14 := position, tokenIndex, depth
						if !_rules[ruleOperator]() {
							goto l15
						}
						if !_rules[ruleSpacing]() {
							goto l15
						}
						if !_rules[ruleValue]() {
							goto l15
						}
						if !_rules[ruleAction3]() {
							goto l15
						}
						goto l14
					l15:
						position, tokenIndex, depth = position14, tokenIndex14, depth14
						if buffer[position] != rune(':') {
							goto l13
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l13
						}
						if !_rules[ruleExpr]() {
							goto l13
						}
						if !_rules[ruleAction4]() {
							goto l13
						}
					}
				l14:
					goto l12
				l13:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if buffer[position] != rune('(') {
						goto l16
					}
					position++
					if !_rules[ruleAction5]() {
						goto l16
					}
					if !_rules[ruleSpacing]() {
						goto l16
					}
					if !_rules[ruleExprs]() {
						goto l16
					}
					if !_rules[ruleAction6]() {
						goto l16
					}
					if !_rules[ruleSpacing]() {
						goto l16
					}
					if buffer[position] != rune(')') {
						goto l16
					}
					position++
					if !_rules[ruleAction7]() {
						goto l16
					}
					goto l12
				l16:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if buffer[position] != rune('N') {
						goto l17
					}
					position++
					if buffer[position] != rune('O') {
						goto l17
					}
					position++
					if buffer[position] != rune('T') {
						goto l17
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l17
					}
					if !_rules[ruleExpr]() {
						goto l17
					}
					if !_rules[ruleAction8]() {
						goto l17
					}
					goto l12
				l17:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if !_rules[ruleValue]() {
						goto l10
					}
					if !_rules[ruleAction9]() {
						goto l10
					}
				}
			l12:
				depth--
				add(ruleExpr, position11)
			}
			return true
		l10:
			position, tokenIndex, depth = position10, tokenIndex10, depth10
			return false
		},
		/* 3 Property <- <(<(([a-z] / [A-Z]) ('_' / [a-z] / [A-Z] / [0-9])* ('.' ([a-z] / [A-Z]) ('_' / [a-z] / [A-Z] / [0-9])*)*)> Action10)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				{
					position20 := position
					depth++
					{
						position21, tokenIndex21, depth21 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l22
						}
						position++
						goto l21
					l22:
						position, tokenIndex, depth = position21, tokenIndex21, depth21
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l18
						}
						position++
					}
				l21:
				l23:
					{
						position24, tokenIndex24, depth24 := position, tokenIndex, depth
						{
							position25, tokenIndex25, depth25 := position, tokenIndex, depth
							if buffer[position] != rune('_') {
								goto l26
							}
							position++
							goto l25
						l26:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l27
							}
							position++
							goto l25
						l27:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l28
							}
							position++
							goto l25
						l28:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l24
							}
							position++
						}
					l25:
						goto l23
					l24:
						position, tokenIndex, depth = position24, tokenIndex24, depth24
					}
				l29:
					{
						position30, tokenIndex30, depth30 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l30
						}
						position++
						{
							position31, tokenIndex31, depth31 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l32
							}
							position++
							goto l31
						l32:
							position, tokenIndex, depth = position31, tokenIndex31, depth31
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l30
							}
							position++
						}
					l31:
					l33:
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							{
								position35, tokenIndex35, depth35 := position, tokenIndex, depth
								if buffer[position] != rune('_') {
									goto l36
								}
								position++
								goto l35
							l36:
								position, tokenIndex, depth = position35, tokenIndex35, depth35
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l37
								}
								position++
								goto l35
							l37:
								position, tokenIndex, depth = position35, tokenIndex35, depth35
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l38
								}
								position++
								goto l35
							l38:
								position, tokenIndex, depth = position35, tokenIndex35, depth35
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l34
								}
								position++
							}
						l35:
							goto l33
						l34:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
						}
						goto l29
					l30:
						position, tokenIndex, depth = position30, tokenIndex30, depth30
					}
					depth--
					add(rulePegText, position20)
				}
				if !_rules[ruleAction10]() {
					goto l18
				}
				depth--
				add(ruleProperty, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 4 Operator <- <(('=' Action11) / ('!' '=' Action12) / ('<' '>' Action13) / ('<' Action14) / ('<' '=' Action15) / ('>' Action16) / ('>' '=' Action17))> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				{
					position41, tokenIndex41, depth41 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l42
					}
					position++
					if !_rules[ruleAction11]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('!') {
						goto l43
					}
					position++
					if buffer[position] != rune('=') {
						goto l43
					}
					position++
					if !_rules[ruleAction12]() {
						goto l43
					}
					goto l41
				l43:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('<') {
						goto l44
					}
					position++
					if buffer[position] != rune('>') {
						goto l44
					}
					position++
					if !_rules[ruleAction13]() {
						goto l44
					}
					goto l41
				l44:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('<') {
						goto l45
					}
					position++
					if !_rules[ruleAction14]() {
						goto l45
					}
					goto l41
				l45:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('<') {
						goto l46
					}
					position++
					if buffer[position] != rune('=') {
						goto l46
					}
					position++
					if !_rules[ruleAction15]() {
						goto l46
					}
					goto l41
				l46:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('>') {
						goto l47
					}
					position++
					if !_rules[ruleAction16]() {
						goto l47
					}
					goto l41
				l47:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
					if buffer[position] != rune('>') {
						goto l39
					}
					position++
					if buffer[position] != rune('=') {
						goto l39
					}
					position++
					if !_rules[ruleAction17]() {
						goto l39
					}
				}
			l41:
				depth--
				add(ruleOperator, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 5 Value <- <(Time / Float / Integer / Bool / String)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !_rules[ruleTime]() {
						goto l51
					}
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if !_rules[ruleFloat]() {
						goto l52
					}
					goto l50
				l52:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if !_rules[ruleInteger]() {
						goto l53
					}
					goto l50
				l53:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if !_rules[ruleBool]() {
						goto l54
					}
					goto l50
				l54:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if !_rules[ruleString]() {
						goto l48
					}
				}
			l50:
				depth--
				add(ruleValue, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 6 Time <- <((<([1-9] [0-9] [0-9] [0-9] '-' [0-9] [0-9] '-' [0-9] [0-9] 'T' [0-9] [0-9] ':' [0-9] [0-9] ':' [0-9] [0-9] 'Z')> Action18) / (<([1-9] [0-9] [0-9] [0-9] '-' [0-9] [0-9] '-' [0-9] [0-9])> Action19))> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					{
						position59 := position
						depth++
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune('-') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune('-') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune('T') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune(':') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune(':') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l58
						}
						position++
						if buffer[position] != rune('Z') {
							goto l58
						}
						position++
						depth--
						add(rulePegText, position59)
					}
					if !_rules[ruleAction18]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					{
						position60 := position
						depth++
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if buffer[position] != rune('-') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if buffer[position] != rune('-') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l55
						}
						position++
						depth--
						add(rulePegText, position60)
					}
					if !_rules[ruleAction19]() {
						goto l55
					}
				}
			l57:
				depth--
				add(ruleTime, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 7 String <- <(BareString / QuotedString)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if !_rules[ruleBareString]() {
						goto l64
					}
					goto l63
				l64:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if !_rules[ruleQuotedString]() {
						goto l61
					}
				}
			l63:
				depth--
				add(ruleString, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 8 BareString <- <(<(([a-z] / [A-Z]) ([a-z] / [A-Z] / [0-9])*)> Action20)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				{
					position67 := position
					depth++
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l69
						}
						position++
						goto l68
					l69:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l65
						}
						position++
					}
				l68:
				l70:
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						{
							position72, tokenIndex72, depth72 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l73
							}
							position++
							goto l72
						l73:
							position, tokenIndex, depth = position72, tokenIndex72, depth72
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l74
							}
							position++
							goto l72
						l74:
							position, tokenIndex, depth = position72, tokenIndex72, depth72
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l71
							}
							position++
						}
					l72:
						goto l70
					l71:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
					}
					depth--
					add(rulePegText, position67)
				}
				if !_rules[ruleAction20]() {
					goto l65
				}
				depth--
				add(ruleBareString, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 9 QuotedString <- <('"' <(!'"' .)*> '"' Action21)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('"') {
					goto l75
				}
				position++
				{
					position77 := position
					depth++
				l78:
					{
						position79, tokenIndex79, depth79 := position, tokenIndex, depth
						{
							position80, tokenIndex80, depth80 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l80
							}
							position++
							goto l79
						l80:
							position, tokenIndex, depth = position80, tokenIndex80, depth80
						}
						if !matchDot() {
							goto l79
						}
						goto l78
					l79:
						position, tokenIndex, depth = position79, tokenIndex79, depth79
					}
					depth--
					add(rulePegText, position77)
				}
				if buffer[position] != rune('"') {
					goto l75
				}
				position++
				if !_rules[ruleAction21]() {
					goto l75
				}
				depth--
				add(ruleQuotedString, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 10 Integer <- <(<([1-9] [0-9]*)> Action22)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				{
					position83 := position
					depth++
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l81
					}
					position++
				l84:
					{
						position85, tokenIndex85, depth85 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l85
						}
						position++
						goto l84
					l85:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
					}
					depth--
					add(rulePegText, position83)
				}
				if !_rules[ruleAction22]() {
					goto l81
				}
				depth--
				add(ruleInteger, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 11 Float <- <(<([1-9] [0-9]* '.' [0-9]+)> Action23)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				{
					position88 := position
					depth++
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l86
					}
					position++
				l89:
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l90
						}
						position++
						goto l89
					l90:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
					}
					if buffer[position] != rune('.') {
						goto l86
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l86
					}
					position++
				l91:
					{
						position92, tokenIndex92, depth92 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l92
						}
						position++
						goto l91
					l92:
						position, tokenIndex, depth = position92, tokenIndex92, depth92
					}
					depth--
					add(rulePegText, position88)
				}
				if !_rules[ruleAction23]() {
					goto l86
				}
				depth--
				add(ruleFloat, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 12 Bool <- <(('t' 'r' 'u' 'e' Action24) / ('f' 'a' 'l' 's' 'e' Action25))> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l96
					}
					position++
					if buffer[position] != rune('r') {
						goto l96
					}
					position++
					if buffer[position] != rune('u') {
						goto l96
					}
					position++
					if buffer[position] != rune('e') {
						goto l96
					}
					position++
					if !_rules[ruleAction24]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if buffer[position] != rune('f') {
						goto l93
					}
					position++
					if buffer[position] != rune('a') {
						goto l93
					}
					position++
					if buffer[position] != rune('l') {
						goto l93
					}
					position++
					if buffer[position] != rune('s') {
						goto l93
					}
					position++
					if buffer[position] != rune('e') {
						goto l93
					}
					position++
					if !_rules[ruleAction25]() {
						goto l93
					}
				}
			l95:
				depth--
				add(ruleBool, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 13 Spacing <- <(Space / Comment)*> */
		func() bool {
			{
				position98 := position
				depth++
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					{
						position101, tokenIndex101, depth101 := position, tokenIndex, depth
						if !_rules[ruleSpace]() {
							goto l102
						}
						goto l101
					l102:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						if !_rules[ruleComment]() {
							goto l100
						}
					}
				l101:
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				depth--
				add(ruleSpacing, position98)
			}
			return true
		},
		/* 14 Comment <- <('#' (!EndOfLine .)*)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != rune('#') {
					goto l103
				}
				position++
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						if !_rules[ruleEndOfLine]() {
							goto l107
						}
						goto l106
					l107:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
					}
					if !matchDot() {
						goto l106
					}
					goto l105
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
				depth--
				add(ruleComment, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 15 Space <- <(' ' / '\t' / EndOfLine)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l111
					}
					position++
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if buffer[position] != rune('\t') {
						goto l112
					}
					position++
					goto l110
				l112:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleEndOfLine]() {
						goto l108
					}
				}
			l110:
				depth--
				add(ruleSpace, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 16 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l116
					}
					position++
					if buffer[position] != rune('\n') {
						goto l116
					}
					position++
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if buffer[position] != rune('\n') {
						goto l117
					}
					position++
					goto l115
				l117:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if buffer[position] != rune('\r') {
						goto l113
					}
					position++
				}
			l115:
				depth--
				add(ruleEndOfLine, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 18 Action0 <- <{ p.reduceAnd() }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 19 Action1 <- <{ p.finalize() }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 20 Action2 <- <{ p.pushOr() }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 21 Action3 <- <{ p.pushOperatorExpr() }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 22 Action4 <- <{ p.pushColonExpr()    }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 23 Action5 <- <{ p.pushNewState() }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 24 Action6 <- <{ p.reduceAnd() }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 25 Action7 <- <{ p.popNewState() }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 26 Action8 <- <{ p.pushNot() }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 27 Action9 <- <{ p.pushKeywordExpr() }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		nil,
		/* 29 Action10 <- <{ p.pushProperty(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 30 Action11 <- <{ p.pushOperator(ast.OpEq)  }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 31 Action12 <- <{ p.pushOperator(ast.OpNeq) }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 32 Action13 <- <{ p.pushOperator(ast.OpNeq) }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 33 Action14 <- <{ p.pushOperator(ast.OpLt)  }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 34 Action15 <- <{ p.pushOperator(ast.OpLe) }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 35 Action16 <- <{ p.pushOperator(ast.OpGt)  }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 36 Action17 <- <{ p.pushOperator(ast.OpGe) }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 37 Action18 <- <{ p.pushTimeValue(time.RFC3339, buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 38 Action19 <- <{ p.pushTimeValue("2006-01-02", buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 39 Action20 <- <{ p.pushStringValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 40 Action21 <- <{ p.pushStringValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 41 Action22 <- <{ p.pushIntegerValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 42 Action23 <- <{ p.pushFloatValue(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 43 Action24 <- <{ p.pushBoolValue(true) }> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 44 Action25 <- <{ p.pushBoolValue(false) }> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
	}
	p.rules = _rules
}
