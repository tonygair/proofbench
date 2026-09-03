package main

import (
	"regexp"
	"strings"
)

// diagnose.go — the DIAGNOSIS LAYER between gnatprove and the model.
//
// Why: a local coder (qwen3-coder) is a reliable EXECUTOR of a fix but cannot
// DIAGNOSE from raw gnatprove output — fed the bare prover dump it loops without
// converging (LLM_IDIOSYNCRASY_REGISTER IR-12). A one-line, class-specific
// diagnosis heals it in one round. This file turns the residual error into that
// diagnosis deterministically — no extra model, fully sovereign, free. It is the
// deterministic-over-reasoning cut applied to proof repair: replace the reasoning
// the model can't do (classify the failure) with code, leaving the model only the
// repair it CAN do.
//
// Mapping (residual signature -> structural hint), grounded in the 2026-06-14
// ceiling battery + the Opus reference bodies that pass each class:
//   - "nonterminating"/Always_Terminates  -> use a bounded `for` loop
//   - "not preserved" + 'Old in the spec    -> 'Loop_Entry frame invariant
//   - "not preserved" (no 'Old)             -> index the invariant by the loop counter
//   - "postcondition ... cannot prove X"     -> add a prefix Loop_Invariant carrying X
// Hints compose; an unrecognised residual returns "" so the raw output still flows.

//   - "loop invariant might fail in first iteration" + nested loops
//                                             -> the TWO-PART inner invariant
//   - "range check might fail ... lower/upper bound" on an accumulation
//                                             -> carry the range bound in the invariant
//
// The last two were added 2026-09-02 from the tier-2 matrix lane (NpTranspose,
// NpTril, NpCumProd): every matrix task failed with "might fail in first
// iteration" because the model wrote a SINGLE-loop invariant shape inside a
// NESTED loop. Stating the two-part shape as prose cracked NpTranspose in one
// round, so it is templated here rather than paid for in prompt length — the
// prose had already saturated (rules present in the prompt were being violated).

var postCannotProveRe = regexp.MustCompile(`(?i)postcondition might fail[,:]?\s*cannot prove\s+([^\n]+)`)

// firstIterationRe catches GNATprove's "loop invariant might fail in first
// iteration", which on a nested loop almost always means the inner invariant
// claims the WHOLE inner range for the current outer row.
var firstIterationRe = regexp.MustCompile(`(?i)loop invariant might fail in first iteration`)

// rangeBoundRe catches an unprovable range check naming the bound it could not
// establish, e.g. "cannot prove lower bound for Result (I - 1) * A (I)".
var rangeBoundRe = regexp.MustCompile(`(?i)range check might fail[,:]?\s*cannot prove (lower|upper) bound for\s+([^\n]+)`)

// nestedLoopRe spots a second `for ... loop` header in the body — the signal
// that a first-iteration failure is the nested-invariant class rather than a
// plain single-loop one.
var nestedLoopRe = regexp.MustCompile(`(?is)\bfor\b[^\n]*\bloop\b.*\bfor\b[^\n]*\bloop\b`)

// diagnoseProofFailure returns a targeted, structural diagnosis for a gnatprove
// proof-failure residual, or "" when no class is recognised. spec is the .ads
// text (used to tell a frame Post — one mentioning 'Old — from an ordinary one);
// body is the candidate body text, used only to tell a nested loop from a single
// one. body may be empty, in which case the nested-loop class degrades to the
// generic first-iteration hint.
func diagnoseProofFailure(proveErrs, spec, body string) string {
	e := strings.ToLower(proveErrs)
	var hints []string

	if strings.Contains(e, "nonterminating") || strings.Contains(e, "always_terminates") {
		hints = append(hints, "TERMINATION: a `while` or plain `loop` cannot be shown to terminate here. "+
			"Rewrite the loop as a bounded `for I in <index range> loop` over the array's range. A for-loop over a "+
			"static range terminates by construction, so the Always_Terminates obligation discharges with no variant. "+
			"A `for` loop may STILL exit early — keep any early `return`/`exit` exactly as is; only change the loop "+
			"HEADER from `while …`/plain `loop` to `for I in A'Range loop`. Do NOT keep a `while` or unbounded `loop`.")
	}

	switch {
	case strings.Contains(e, "not preserved") && strings.Contains(spec, "'Old"):
		hints = append(hints, "FRAME / IN-PLACE: the postcondition compares against `'Old`, but a loop invariant may NOT use `'Old` — "+
			"use `'Loop_Entry` (the array as it was when the loop began). Assert TWO invariants every iteration: (1) the portion already "+
			"mutated equals its final target written in terms of `'Loop_Entry`; (2) the portion not yet touched still equals its `'Loop_Entry` "+
			"value — the FRAME. The Post cannot carry without the untouched-frame invariant.")
	case strings.Contains(e, "not preserved"):
		hints = append(hints, "NON-INDUCTIVE INVARIANT: the invariant holds initially but not after an arbitrary iteration. Strengthen it so that, "+
			"assumed at the top of the body, it still holds at the bottom — usually by indexing the asserted property by the loop counter "+
			"(state what holds for the prefix `A'First .. I`).")
	}

	if m := postCannotProveRe.FindStringSubmatch(proveErrs); m != nil {
		hints = append(hints, "MISSING PREFIX INVARIANT: the postcondition `"+strings.TrimSpace(m[1])+"` cannot be carried out of the loop. "+
			"Add a `pragma Loop_Invariant` at the END of the loop body stating this property for the prefix scanned so far (quantify over "+
			"`A'First .. I`, nesting one invariant per loop for nested loops), so at loop exit (I = A'Last) the invariant IS the postcondition.")
	}

	// FIRST-ITERATION class. On a nested loop this is nearly always the inner
	// invariant claiming the whole inner range for the CURRENT outer row, which
	// is false when only one element of that row has been written. Measured on
	// the tier-2 matrix lane 2026-09-02: every matrix task failed exactly here,
	// and the two-part shape below proved NpTranspose in one round.
	if firstIterationRe.MatchString(proveErrs) {
		if nestedLoopRe.MatchString(body) {
			hints = append(hints, "NESTED-LOOP INVARIANT — TWO PARTS, NOT ONE: the inner invariant must state exactly what has been written "+
				"at that moment: (a) every EARLIER outer row, COMPLETELY, and (b) the CURRENT outer row only UP TO the current inner index. "+
				"Claiming the whole inner range for the current row is FALSE IN THE FIRST INNER ITERATION — that is what the prover is "+
				"reporting. Shape: `pragma Loop_Invariant ((for all K in A'First (1) .. I - 1 => (for all L in A'Range (2) => <done>)) "+
				"and then (for all L in A'First (2) .. J => <done for row I>));` at the END of the INNER body, PLUS an outer "+
				"`pragma Loop_Invariant` at the END of the OUTER body stating rows A'First (1) .. I are complete — without the outer one, "+
				"part (a) has nothing to stand on in the next pass. Where the element written depends on a condition (a mask), state BOTH "+
				"regions in BOTH parts, reusing the Post's own test verbatim.")
		} else {
			hints = append(hints, "INVARIANT FALSE ON ENTRY: the invariant does not hold the first time it is reached. State it over the prefix "+
				"ALREADY written (`A'First .. I`), not over the whole range — at the first iteration only one element exists, so any claim "+
				"about the untouched remainder is false. Place it at the END of the loop body, after the work.")
		}
	}

	// RANGE-BOUND class. The prover is not doubting the algorithm; it lacks a
	// fact about how large the running value can be BEFORE the next operation.
	// NpCumProd failed on this identically in three consecutive rounds.
	if m := rangeBoundRe.FindStringSubmatch(proveErrs); m != nil {
		hints = append(hints, "MISSING RANGE FACT IN THE INVARIANT: the prover cannot establish the "+strings.ToLower(m[1])+
			" bound for `"+strings.TrimSpace(m[2])+"`. It is not questioning your algorithm — at that point nothing tells it the running "+
			"value is still small enough for the next operation to stay in range. Add an explicit TWO-SIDED bound on every element written "+
			"so far to the loop invariant, using the bounding constant the preamble declares (it is declared because this proof needs it), "+
			"e.g. `(for all K in A'First .. I => Result (K) in -Max_Bound .. Max_Bound)`, AND keep the equality conjunct that carries the "+
			"postcondition. State both signs — a negative element is exactly the lower bound it cannot prove.")
	}

	// DECOMPOSITION comes first when the Post is conjunctive: knowing WHICH of
	// several theorems was rejected shrinks the question before any of the
	// structural hints above are acted on. Deterministic — see decompose.go.
	if d := decomposeHint(spec, proveErrs); d != "" {
		hints = append([]string{d}, hints...)
	}

	// THE DERIVED INVARIANT leads everything. Where the residual is a missing
	// prefix invariant for a quantified Post — the shape qwen has been measured
	// four separate ways unable to produce, or even to copy when handed it — we
	// do not ask for it: synthesise.go derives it from the Post and hands over
	// the literal Ada. Only offered when that is actually the complaint, so a
	// range-check or frame failure still gets its own diagnosis first.
	if postCannotProveRe.MatchString(proveErrs) || firstIterationRe.MatchString(proveErrs) ||
		strings.Contains(e, "not preserved") {
		if inv := synthesisedInvariantHint(spec, body); inv != "" {
			hints = append([]string{inv}, hints...)
		}
	}

	if len(hints) == 0 {
		return ""
	}
	return "DIAGNOSIS — act on this FIRST, it is the specific cause:\n- " + strings.Join(hints, "\n- ")
}
