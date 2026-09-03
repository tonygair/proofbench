package main

import (
	"fmt"
	"regexp"
	"strings"
)

// synthesise.go — DERIVE THE PREFIX LOOP INVARIANT FROM THE POSTCONDITION.
//
// The sovereign forge's oldest named limit (fact_forge_qwen_loop_invariant_gap,
// 2026-07-07): qwen cannot emit an inductive prefix loop-invariant for a
// quantified Post — and, critically, fails **even when the exact invariant is
// given to it**. Confirmed three more ways on 2026-09-02: raw prover output fed
// back 8× produced byte-identical bodies; Kevin's named diagnosis did not move
// it; and a reasoning model (deepseek-r1) handed it a correct, specific fix which
// it still did not apply.
//
// The July finding is the load-bearing one: the model fails to TYPE the
// invariant, not to understand it. So do not ask it. Derive the invariant here
// and hand it over as literal Ada.
//
// This is possible because the shape is a REWRITE, not a judgement:
//
//	Post => (for all I in A'Range => P (I))
//	  =>   pragma Loop_Invariant (for all K in A'First .. I => P (K));
//
// The postcondition says a property holds over the WHOLE range at exit; the
// invariant says it holds over the PREFIX SCANNED SO FAR. Restricting the
// quantifier's range is mechanical. decompose.go already parses the Post into
// conjuncts, so each conjunct yields its own prefix clause.
//
// RECOGNITION IS NOT OUR JOB. Whether the derived invariant is actually
// inductive is decided by GNATprove — free, certain, and better than any model.
// This file only proposes; the prover disposes. A proposal that does not
// discharge costs one round and is discarded.
//
// Conservative by construction: a conjunct whose quantifier we cannot parse
// yields nothing rather than a guess, and nothing here edits code or weakens a
// Post — the caller decides what to do with the text.

// forAllRe matches a universally quantified conjunct and captures the loop
// variable, the range expression, and the body of the quantifier.
//
//	(for all I in A'Range => <body>)
//	(for all I in A'First .. A'Last => <body>)
var forAllRe = regexp.MustCompile(`(?is)^\(?\s*for\s+all\s+([A-Za-z_]\w*)\s+in\s+(.+?)\s*=>\s*(.*?)\)?\s*$`)

// rangeAttrRe recognises `X'Range`, which restricts to `X'First .. <counter>`.
var rangeAttrRe = regexp.MustCompile(`(?i)^([A-Za-z_]\w*)\s*'\s*Range$`)

// explicitRangeRe recognises `<low> .. <high>`.
var explicitRangeRe = regexp.MustCompile(`(?is)^(.+?)\s*\.\.\s*(.+)$`)

// prefixClause turns ONE universally quantified conjunct into its prefix form,
// quantified up to the loop counter. Returns "" when the conjunct is not a
// universal quantifier we can restrict — a guess is worse than silence.
func prefixClause(conjunct, counter string) string {
	c := strings.TrimSpace(conjunct)
	m := forAllRe.FindStringSubmatch(c)
	if m == nil {
		return ""
	}
	loopVar, rangeExpr, body := strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), strings.TrimSpace(m[3])
	if body == "" {
		return ""
	}
	// A nested quantifier inside the body is a matrix/two-index shape; the prefix
	// rewrite for those is the TWO-PART invariant (see diagnose.go), not this
	// one-dimensional restriction. Refuse rather than emit a wrong shape.
	if strings.Contains(strings.ToLower(body), "for all") || strings.Contains(strings.ToLower(body), "for some") {
		return ""
	}

	var lo string
	switch {
	case rangeAttrRe.MatchString(rangeExpr):
		lo = rangeAttrRe.FindStringSubmatch(rangeExpr)[1] + "'First"
	case explicitRangeRe.MatchString(rangeExpr):
		lo = strings.TrimSpace(explicitRangeRe.FindStringSubmatch(rangeExpr)[1])
	default:
		return ""
	}

	// Keep the author's own loop variable name so the body needs no rewriting.
	return fmt.Sprintf("(for all %s in %s .. %s => %s)", loopVar, lo, counter, body)
}

// loopCounterRe finds the counter of the FIRST `for` loop in a body, which is
// the variable a prefix invariant must be stated against.
var loopCounterRe = regexp.MustCompile(`(?is)\bfor\s+([A-Za-z_]\w*)\s+in\b`)

// loopCounterOf returns the first loop parameter in the body, or "I" when the
// body has no readable loop header (the conventional name in this shop's specs).
func loopCounterOf(body string) string {
	if m := loopCounterRe.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return "I"
}

// synthesisePrefixInvariant derives the prefix Loop_Invariant carrying a
// quantified postcondition, as literal Ada ready to paste at the END of the loop
// body. Returns "" when no conjunct can be restricted mechanically.
//
// It does NOT decide whether the invariant is inductive — GNATprove does. The
// point is to stop asking a model for a text it has been measured, four times,
// unable to produce or even copy.
func synthesisePrefixInvariant(spec, body string) string {
	post := extractPost(spec)
	if post == "" {
		return ""
	}
	counter := loopCounterOf(body)
	var clauses []string
	for _, conj := range splitConjuncts(post) {
		if cl := prefixClause(conj, counter); cl != "" {
			clauses = append(clauses, cl)
		}
	}
	if len(clauses) == 0 {
		return ""
	}
	return "pragma Loop_Invariant\n  (" + strings.Join(clauses, "\n   and then ") + ");"
}

// synthesisedInvariantHint wraps the derived invariant as an instruction. The
// text is handed over verbatim precisely because the model has been measured
// unable to author it: its job is to place it, not to invent it.
func synthesisedInvariantHint(spec, body string) string {
	inv := synthesisePrefixInvariant(spec, body)
	if inv == "" {
		return ""
	}
	return "THE INVARIANT, DERIVED FOR YOU — do not invent your own, and do not reword this one. " +
		"It is the postcondition restricted to the prefix scanned so far, which is what carries the Post out of " +
		"the loop. Paste it VERBATIM as the LAST statement inside the loop body, after the work of the iteration:\n\n" +
		inv + "\n\n" +
		"Change nothing else. If the prover still objects, it will name what is missing and that is a SEPARATE fact " +
		"to add (a range bound, a frame clause) — not a reason to rewrite this clause."
}
