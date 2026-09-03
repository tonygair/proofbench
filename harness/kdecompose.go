package main

import (
	"fmt"
	"regexp"
	"strings"
)

// decompose.go — DETERMINISTIC DECOMPOSITION OF A CONJUNCTIVE POSTCONDITION.
//
// A Post of the form `A and then B and then C` is three theorems, and the model
// must satisfy all of them with one invariant. The commonest way it fails is to
// establish one conjunct and forget another — a failure the Wu prose already
// names, and therefore one we have been paying for in prompt length.
//
// But the split is PARSING, not judgement: `and`/`and then` at parenthesis depth
// zero separates the conjuncts, and GNATprove's residual NAMES the conjunct it
// could not prove ("postcondition might fail, cannot prove Occurrences (Result,
// X) = Occurrences (A, X)"). So the whole decomposition — how many theorems,
// which one failed, which ones already discharge — is decidable. Never train
// what you can decide.
//
// The point is not to tell the model the answer. It is to shrink the question:
// repair the clause supporting conjunct k, and leave the conjuncts that already
// discharge alone. Measured motivation (2026-09-02, tier-2 numpy lane): NpSort
// carries `sortedness AND multiset-preservation` and failed on the multiset half
// in every round, while the model kept rewriting the whole body.
//
// Nothing here edits code or weakens a Post. The spec is fixed; this only reads
// it.

// postRe pulls the Post aspect's expression out of a spec. It deliberately stops
// at the aspect terminator (`;` at depth 0), handled by the depth walk below.
var postRe = regexp.MustCompile(`(?is)\bPost\s*=>\s*`)

// cannotProveRe names the expression GNATprove could not establish.
var cannotProveRe = regexp.MustCompile(`(?is)cannot prove\s+(.+?)(?:\n|$)`)

// extractPost returns the text of the Post aspect's expression, or "" if the
// spec has none. It walks parenthesis depth so a Post containing `;`-free nested
// aspects or quantifiers is not cut short.
func extractPost(spec string) string {
	loc := postRe.FindStringIndex(spec)
	if loc == nil {
		return ""
	}
	rest := spec[loc[1]:]
	depth := 0
	for i, r := range rest {
		switch r {
		case '(':
			depth++
		case ')':
			// A closing paren at depth 0 ends the enclosing aspect list.
			if depth == 0 {
				return strings.TrimSpace(rest[:i])
			}
			depth--
		case ';':
			if depth == 0 {
				return strings.TrimSpace(rest[:i])
			}
		}
	}
	return strings.TrimSpace(rest)
}

// splitConjuncts splits a boolean expression into its top-level conjuncts,
// respecting parentheses and string literals. `and then` and `and` both
// separate; `or`/`or else` do NOT — a disjunctive Post is returned whole,
// because splitting it would be a lie about what must hold.
func splitConjuncts(expr string) []string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}
	var (
		parts   []string
		start   int
		depth   int
		inQuote bool
	)
	lower := strings.ToLower(expr)
	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '"':
			inQuote = !inQuote
			continue
		case '(':
			if !inQuote {
				depth++
			}
			continue
		case ')':
			if !inQuote {
				depth--
			}
			continue
		}
		if inQuote || depth != 0 {
			continue
		}
		// only match `and` on a word boundary at depth 0
		if strings.HasPrefix(lower[i:], "and") && isBoundary(lower, i-1) {
			after := i + len("and")
			if strings.HasPrefix(lower[after:], " then") {
				after += len(" then")
			}
			if !isBoundary(lower, after) {
				continue
			}
			parts = append(parts, strings.TrimSpace(expr[start:i]))
			start = after
			i = after - 1
		}
	}
	parts = append(parts, strings.TrimSpace(expr[start:]))
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isBoundary(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return true
	}
	c := s[i]
	return !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9')
}

// normaliseSpace collapses all runs of whitespace so a conjunct reflowed across
// lines still matches the prover's single-line rendering of it.
func normaliseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// failingConjunct returns the 1-based index of the conjunct GNATprove named as
// unprovable, and the conjunct text. It returns 0 when the residual names no
// expression or the expression matches none of the conjuncts.
func failingConjunct(proveErrs string, conjuncts []string) (int, string) {
	m := cannotProveRe.FindStringSubmatch(proveErrs)
	if m == nil {
		return 0, ""
	}
	named := normaliseSpace(strings.TrimSpace(m[1]))
	named = strings.TrimSuffix(named, ".")
	if named == "" {
		return 0, ""
	}
	// Exact-ish match first, then containment either way: the prover may print a
	// sub-expression of the conjunct, or the conjunct wrapped in parentheses.
	for i, c := range conjuncts {
		n := normaliseSpace(c)
		if n == named {
			return i + 1, c
		}
	}
	for i, c := range conjuncts {
		n := normaliseSpace(c)
		if strings.Contains(n, named) || strings.Contains(named, n) {
			return i + 1, c
		}
	}
	return 0, ""
}

// decomposeHint turns a conjunctive Post plus a proof residual into a hint that
// SHRINKS the question: how many theorems there are, which one the prover
// rejected, and an instruction to leave the discharged ones alone. Returns ""
// when the Post is not conjunctive or nothing can be attributed — in which case
// the caller's other diagnoses still stand on their own.
func decomposeHint(spec, proveErrs string) string {
	post := extractPost(spec)
	if post == "" {
		return ""
	}
	conjuncts := splitConjuncts(post)
	if len(conjuncts) < 2 {
		return ""
	}
	idx, text := failingConjunct(proveErrs, conjuncts)

	var b strings.Builder
	fmt.Fprintf(&b, "DECOMPOSITION: this postcondition is %d SEPARATE THEOREMS joined by `and`. "+
		"They must ALL hold, and your loop invariant needs a clause carrying EACH one — establishing some and "+
		"forgetting the rest is the commonest way this task is failed.\n", len(conjuncts))
	for i, c := range conjuncts {
		mark := "  "
		if i+1 == idx {
			mark = "->"
		}
		fmt.Fprintf(&b, "   %s (%d) %s\n", mark, i+1, truncateConjunct(normaliseSpace(c)))
	}
	if idx > 0 {
		fmt.Fprintf(&b, "   The prover rejected ONLY conjunct (%d): `%s`. The others are discharging — do NOT rewrite "+
			"the parts of the body or the invariant that serve them. Add or strengthen exactly the invariant clause that "+
			"carries conjunct (%d), and change nothing else.", idx, truncateConjunct(normaliseSpace(text)), idx)
	} else {
		b.WriteString("   Give the invariant one clause per conjunct, in the same order, so that at loop exit the " +
			"conjunction of the clauses IS the postcondition.")
	}
	return b.String()
}

func truncateConjunct(s string) string {
	const max = 160
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
