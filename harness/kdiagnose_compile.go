package main

import (
	"regexp"
	"strings"
)

// diagnose_compile.go — the DIAGNOSIS LAYER between the COMPILER and the model.
//
// Sibling of diagnose.go, which does the same job for gnatprove residuals. The
// same finding motivates both (LLM_IDIOSYNCRASY_REGISTER IR-12): a local coder
// REPAIRS well but DIAGNOSES badly, and fed a raw tool dump it loops without
// converging. Measured again 2026-09-02 on the unfettered tier-2 loop: eight
// repair attempts per task with the raw judge output fed straight back produced
// byte-identical judge timings every round — the model emitted the same body
// each time. Raw output is not feedback; a named class is.
//
// Every class below was emitted by qwen3-coder:30b on the tier-2 numpy lane
// (rounds A-E, 2026-09-02) and rejected by gnatchop/gcc. They are Ada legality
// faults, not proof faults, so they are decidable by inspection — exactly the
// work that belongs in code rather than in prompt length. Several were emitted
// even though the prompt already forbade them: at ~15KB the prose had saturated,
// which is the argument for templating them here.
//
// An unrecognised diagnostic returns "" so the raw compiler output still flows.

var (
	// pragma Loop_Invariant outside the statements of a loop (typically put in
	// the declarative part, before `begin`).
	invariantPlacementRe = regexp.MustCompile(`(?i)pragma "?Loop_Invariant"? must appear immediately within the statements of a loop`)

	// A helper already declared in the preamble re-declared in <vc-helpers>.
	duplicateBodyRe = regexp.MustCompile(`(?i)duplicate body for "([^"]+)"`)

	// `J := J;` and friends — a loop parameter is constant.
	loopParamAssignRe = regexp.MustCompile(`(?i)assignment to loop parameter not allowed`)

	// A quantified expression used as a VALUE, e.g. `X = (for some ...)`.
	boolExpectedRe = regexp.MustCompile(`(?i)expected type "Standard\.Boolean"`)

	// The body's unit name does not match the spec it must complete.
	missingSpecRe = regexp.MustCompile(`(?i)file "([^"]+)\.ads" not found`)

	// Plain syntax slips, overwhelmingly the commonest single class.
	missingTokenRe = regexp.MustCompile(`(?i)missing "([^"]+)"`)

	// An aspect written where none may appear (usually a stray `with` clause
	// hung off a declaration inside the body).
	aspectNotAllowedRe = regexp.MustCompile(`(?i)aspect specifications not allowed here`)

	// `others` used outside a whole-object initialiser — e.g. inside a pragma.
	othersRe = regexp.MustCompile(`(?i)("others"|OTHERS).*(not allowed|illegal|cannot)`)
)

// diagnoseCompileFailure returns a targeted, structural diagnosis for a compiler
// or gnatchop rejection, or "" when no class is recognised. compileErrs is the
// raw tool output; body is the candidate body text (used to sharpen a couple of
// classes) and may be empty.
func diagnoseCompileFailure(compileErrs, body string) string {
	var hints []string

	if invariantPlacementRe.MatchString(compileErrs) {
		hints = append(hints, "LOOP_INVARIANT MISPLACED: a `pragma Loop_Invariant` may appear ONLY immediately inside the STATEMENTS of a loop — "+
			"never in the declarative part (before `begin`), never before the loop, never after `end loop`. If you meant to initialise the "+
			"whole array, that is a plain STATEMENT after `begin` and carries no pragma: `Result := (others => 0);`. Move the invariant to the "+
			"END of the loop body, after the work of the iteration.")
	}

	if m := duplicateBodyRe.FindStringSubmatch(compileErrs); m != nil {
		hints = append(hints, "DUPLICATE HELPER: `"+strings.TrimSpace(m[1])+"` is ALREADY declared for you in the preamble — that part of the file is "+
			"fixed and you must not restate it. Delete your re-declaration from <vc-helpers> and simply CALL it. Only add a helper that does "+
			"not already exist.")
	}

	if loopParamAssignRe.MatchString(compileErrs) {
		hints = append(hints, "ASSIGNMENT TO A LOOP PARAMETER: the control variable of a `for` loop is CONSTANT inside the loop and can never be "+
			"assigned. If you need an index you can move, declare your own variable before the loop and assign to that; if you were trying to "+
			"skip iterations, restructure the loop or use an inner condition instead.")
	}

	if boolExpectedRe.MatchString(compileErrs) {
		hints = append(hints, "QUANTIFIER USED AS A VALUE: a quantified expression `(for all ...)` / `(for some ...)` is a BOOLEAN. It is not an "+
			"element, not an index and not a number, so it can never be an operand of `=`, `+` or any operator expecting a value. Put the "+
			"comparison INSIDE the quantifier: not `Result (K) = (for some L in R => ...)` but `(for some L in R => Result (K) = A (L))`.")
	}

	if othersRe.MatchString(compileErrs) {
		hints = append(hints, "`others` OUT OF PLACE: `others` is legal only in a whole-object aggregate assigned to the object itself "+
			"(`Result := (others => 0);`). It may never appear inside a `pragma Loop_Invariant`, a `pragma Assert`, or any other pragma or "+
			"expression. To say every element holds a value in a pragma, quantify: `(for all I in Result'Range => Result (I) = 0)`.")
	}

	if m := missingSpecRe.FindStringSubmatch(compileErrs); m != nil {
		hints = append(hints, "UNIT NAME MISMATCH: the compiler is looking for spec `"+strings.TrimSpace(m[1])+".ads` because that is the unit your "+
			"body names. The body must complete the package EXACTLY as the fixed spec declares it — copy the package name from the spec "+
			"verbatim, including underscores and capitalisation. Do not invent, shorten or re-case it.")
	}

	if aspectNotAllowedRe.MatchString(compileErrs) {
		hints = append(hints, "ASPECT IN AN ILLEGAL POSITION: a `with Pre =>`/`with Post =>`/`with Subprogram_Variant =>` clause may only hang off a "+
			"DECLARATION, and a subprogram BODY in the body part must not repeat the contract already given in the spec. Drop the aspect from "+
			"the body; the contract lives in the fixed spec and is already in force.")
	}

	// Reported last: a bare syntax slip is usually a symptom of one of the above
	// rather than the cause, so the structural hints should be read first.
	if m := missingTokenRe.FindStringSubmatch(compileErrs); m != nil {
		hints = append(hints, "SYNTAX — MISSING `"+strings.TrimSpace(m[1])+"`: the parser stopped at a missing token, so nothing after that point was even "+
			"read. Fix it before anything else: check that every declaration ends in `;`, every parameter is separated by `,`, and that the "+
			"parentheses balance on the line the compiler names. Emit the sections again in full; do not patch around the error.")
	}

	if len(hints) == 0 {
		return ""
	}
	return "DIAGNOSIS — the body did not COMPILE. Act on this FIRST, it is the specific cause:\n- " + strings.Join(hints, "\n- ")
}
