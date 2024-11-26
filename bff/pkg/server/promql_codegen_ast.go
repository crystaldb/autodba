package server

import (
	"fmt"
	"sort"
	"strings"
)

// Node interface represents a node in the AST.
type Node interface {
	String() string
}

// Selector represents a simple label selector in PromQL.
type Selector struct {
	Metric string
	Labels map[string]string
}

func (s *Selector) String() string {
	var labelParts []string

	// sort the keys to ensure consistent output
	keys := make([]string, 0, len(s.Labels))
	for k := range s.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Add all other labels
	for _, k := range keys {
		labelParts = append(labelParts, fmt.Sprintf(`%s=~"%s"`, k, s.Labels[k]))
	}

	return fmt.Sprintf(`%s{%s}`, s.Metric, strings.Join(labelParts, ","))
}

// Aggregation represents an aggregation function in PromQL.
type Aggregation struct {
	Func string
	By   []string
	Expr Node
}

func (a *Aggregation) String() string {
	byClause := ""
	if len(a.By) > 0 {
		byClause = fmt.Sprintf(" by(%s)", strings.Join(a.By, ", "))
	}
	return fmt.Sprintf(`%s%s (%s)`, a.Func, byClause, a.Expr.String())
}

// FunctionCall represents a PromQL function call.
type FunctionCall struct {
	Func         string
	Args         []Node
	TimeInterval *LiteralInt // Duration of the time range
	TimeStep     *LiteralInt // Step size for the range query
}

func (f *FunctionCall) String() string {
	var args []string
	for _, arg := range f.Args {
		args = append(args, arg.String())
	}
	if f.TimeInterval != nil {
		if f.TimeStep != nil {
			return fmt.Sprintf(`%s(%s[%s:%ss])`, f.Func, strings.Join(args, ", "), f.TimeInterval, f.TimeStep)
		}
		return fmt.Sprintf(`%s(%s[%s:])`, f.Func, strings.Join(args, ", "), f.TimeInterval)
	}
	return fmt.Sprintf(`%s(%s)`, f.Func, strings.Join(args, ", "))
}

// LabelReplace represents the label_replace function in PromQL.
type LabelReplace struct {
	Expr        Node
	DstLabel    string
	Replacement string
	SrcLabel    string
	Regex       string
}

func (lr *LabelReplace) String() string {
	return fmt.Sprintf(`label_replace(%s, "%s", "%s", "%s", "%s")`,
		lr.Expr.String(), lr.DstLabel, lr.Replacement, lr.SrcLabel, lr.Regex)
}

// Topk represents the topk function in PromQL.
type Topk struct {
	Limit int
	Expr  Node
}

func (t *Topk) String() string {
	return fmt.Sprintf("topk(%d, %s)", t.Limit, t.Expr.String())
}

// Bottomk represents the bottomk function in PromQL.
type Bottomk struct {
	Limit int
	Expr  Node
}

func (b *Bottomk) String() string {
	return fmt.Sprintf("bottomk(%d, %s)", b.Limit, b.Expr.String())
}

// SortDesc represents the sort_desc function in PromQL.
type SortDesc struct {
	Expr Node
}

func (s *SortDesc) String() string {
	return fmt.Sprintf("sort_desc(%s)", s.Expr.String())
}

// LiteralInt represents a literal integer value in PromQL.
type LiteralInt struct {
	Value string
}

// String returns the string representation of the literalInt value.
func (l *LiteralInt) String() string {
	return l.Value
}

// LiteralString represents a literal striing value in PromQL.
type LiteralString struct {
	Value string
}

// String returns the string representation of the literalString value.
func (l *LiteralString) String() string {
	return fmt.Sprintf("\"%s\"", l.Value)
}

// BinaryExpr represents a binary expression in PromQL.
type BinaryExpr struct {
	Op    string   // The binary operation (e.g., *, +, -, /, etc.)
	LHS   Node     // The left-hand side expression
	RHS   Node     // The right-hand side expression
	On    []string // The labels to match on (for on(label) operator)
	Group string   // The grouping modifier, e.g., "group_left", "group_right", if applicable
	Bool  bool     // Whether to use the bool modifier (for comparison operators)
}

// String representation for PromQL (example implementation)
func (b *BinaryExpr) String() string {
	var onClause string
	if len(b.On) > 0 {
		onClause = fmt.Sprintf(" on(%s)", strings.Join(b.On, ","))
	}
	var groupClause string = ""
	if b.Group != "" {
		groupClause = fmt.Sprintf(" %s", b.Group)
	}
	boolModifier := ""
	if b.Bool {
		boolModifier = "bool "
	}
	return fmt.Sprintf("%s %s%s%s %s%s", b.LHS.String(), b.Op, onClause, groupClause, boolModifier, b.RHS.String())
}
