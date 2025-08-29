package format

import (
	"strings"

	"codeberg.org/bvaudour/kcp/pkgbuild/info"
	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"codeberg.org/bvaudour/kcp/pkgbuild/standard"
	"mvdan.cc/sh/v3/syntax"
)

type quoting struct {
	fromQuote      string
	quotable       bool
	unquotable     bool
	singlequotable bool
}

func (q quoting) join(other quoting) quoting {
	return quoting{
		quotable:       q.quotable && other.quotable,
		unquotable:     q.unquotable && other.unquotable,
		singlequotable: q.singlequotable && other.singlequotable,
	}
}

func (q quoting) isNil() bool {
	return !q.quotable && !q.unquotable && !q.singlequotable
}

type joinQuoting struct {
	startIndex int
	endIndex   int
	quoting
}

func getQuoting(p syntax.WordPart) (q quoting) {
	switch x := p.(type) {
	case *syntax.Lit:
		q.quotable, q.unquotable = checkConvert(x.Value, q.fromQuote)
		if q.quotable {
			q.singlequotable = canSingleQuote(x.Value)
		}
	case *syntax.SglQuoted:
		q = quoting{
			fromQuote:      `'`,
			quotable:       true,
			singlequotable: true,
			unquotable:     canUnquote(x.Value),
		}
	case *syntax.DblQuoted:
		q = quoting{
			fromQuote:      `"`,
			quotable:       true,
			singlequotable: true,
			unquotable:     true,
		}
		for _, pp := range x.Parts {
			if y, ok := pp.(*syntax.Lit); ok {
				if !canSingleQuote(y.Value) {
					q.singlequotable = false
				}
				if !canUnquote(y.Value) {
					q.unquotable = false
				}
			} else {
				q.singlequotable, q.unquotable = false, false
			}
			if !q.singlequotable && !q.unquotable {
				break
			}
		}
	default:
		q.quotable = true
	}
	return
}

func minimizeQuotingSequence(quotings []quoting) []joinQuoting {
	n := len(quotings)
	if n == 0 {
		return nil
	}

	// Step 1: Pre-compute all possible joins and their validity.
	// precomputedJoins[i][j] holds the result of joining elements from i to j.
	precomputedJoins := make([][]quoting, n)
	for i := range precomputedJoins {
		precomputedJoins[i] = make([]quoting, n)
	}

	for i := 0; i < n; i++ {
		precomputedJoins[i][i] = quotings[i]
		for j := i + 1; j < n; j++ {
			precomputedJoins[i][j] = precomputedJoins[i][j-1].join(quotings[j])
		}
	}

	// Step 2: Dynamic Programming to find the optimal partition.
	// dp[i] stores the maximum score (sum of squares of group lengths) for the first i elements.
	// path[i] stores the start index of the last group in the optimal partition for the first i elements.
	dp := make([]int, n+1)
	path := make([]int, n+1)

	for i := 1; i <= n; i++ {
		dp[i] = -1 // Use -1 to indicate that no valid partition has been found yet.
		for j := 1; j <= i; j++ {
			startIdx := j - 1
			endIdx := i - 1

			// Check if the group from startIdx to endIdx is valid (not nil).
			if !precomputedJoins[startIdx][endIdx].isNil() {
				groupSize := i - startIdx
				// The score of a partition is the sum of the squares of its group lengths.
				// This scoring system favors larger groups.
				score := dp[startIdx] + groupSize*groupSize

				if score > dp[i] {
					dp[i] = score
					path[i] = startIdx
				}
			}
		}
	}

	// Step 3: Reconstruct the solution by backtracking from the end.
	var results []joinQuoting
	curr := n
	for curr > 0 {
		prev := path[curr]
		results = append(results, joinQuoting{
			startIndex: prev,
			endIndex:   curr - 1,
			quoting:    precomputedJoins[prev][curr-1],
		})
		curr = prev
	}

	// Reverse the results slice because we built it backwards.
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results
}

func paramExpToLong(n *syntax.ParamExp) (diff *position.PosDiff) {
	if n.Short {
		end := n.End()
		diff = &position.PosDiff{
			Col:    1,
			Offset: 1,
		}

		n.Short = false
		diff.Update(n.Param)
		n.Rbrace = n.Param.ValueEnd

		diff.Col++
		diff.Offset++
		diff.From = end
		diff.Ignore = n
	}

	return
}

func transformParams(word *syntax.Word) {
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.ParamExp:
			diff := paramExpToLong(p)
			diff.Update(word)
		case *syntax.DblQuoted:
			for _, dpart := range p.Parts {
				if dp, ok := dpart.(*syntax.ParamExp); ok {
					diff := paramExpToLong(dp)
					diff.Update(word)
				}
			}
		}
	}
}

func stringOf(n syntax.Node) string {
	var sb strings.Builder
	pr := syntax.NewPrinter()
	pr.Print(&sb, n)
	return sb.String()
}

func allStringOf(parts []syntax.WordPart) string {
	var sb strings.Builder
	for _, p := range parts {
		sb.WriteString(stringOf(p))
	}
	return sb.String()
}

func formatWord(word *syntax.Word, preferQuote bool) {
	if word == nil {
		return
	}
	transformParams(word)

	quotings := make([]quoting, len(word.Parts))
	for i, part := range word.Parts {
		quotings[i] = getQuoting(part)
	}
	joins := minimizeQuotingSequence(quotings)

	var sb strings.Builder
	for _, j := range joins {
		var toQuote string
		if !j.unquotable || (preferQuote && j.quotable) {
			if j.singlequotable {
				toQuote = `'`
			} else {
				toQuote = `"`
			}
		}
		sb.WriteString(toQuote)
		for i := j.startIndex; i <= j.endIndex; i++ {
			fromQuote := quotings[i].fromQuote
			var s string
			switch part := word.Parts[i].(type) {
			case *syntax.Lit:
				s = part.Value
			case *syntax.SglQuoted:
				s = part.Value
			case *syntax.DblQuoted:
				s = allStringOf(part.Parts)
			default:
				s = stringOf(part)
			}
			newWord, _ := bestConvert(s, fromQuote, toQuote)
			sb.WriteString(newWord)
		}
		sb.WriteString(toQuote)
	}

	parser := syntax.NewParser(syntax.KeepComments(true))

	f, err := parser.Parse(strings.NewReader(sb.String()), "")
	if err != nil || len(f.Stmts) == 0 {
		return
	}

	expr, ok := f.Stmts[0].Cmd.(*syntax.CallExpr)
	if !ok || len(expr.Args) == 0 || len(expr.Args[0].Parts) == 0 {
		return
	}

	newWord := expr.Args[0]
	diff := position.Diff(newWord.Pos(), word.Pos())
	diff.Update(newWord)

	word.Parts = make([]syntax.WordPart, len(newWord.Parts))
	copy(word.Parts, newWord.Parts)
}

func formatWordRecursive(node *info.NodeInfo, preferQuote bool) *info.NodeInfo {
	syntax.Walk(node.Stmt, func(n syntax.Node) bool {
		if word, ok := n.(*syntax.Word); ok && word != nil {
			oldEnd := word.End()
			formatWord(word, preferQuote)
			newEnd := word.End()
			diff := position.Diff(oldEnd, newEnd)
			diff.Ignore = word
			diff.Update(node.Stmt)
		}
		return true
	})

	return node
}

func FormatWords(node *info.NodeInfo) *info.NodeInfo {
	preferQuote := node.Type == info.Function ||
		!standard.IsStandardVariable(node.Name) ||
		standard.IsQuotedVariable(node.Name)

	return formatWordRecursive(node, preferQuote)
}
