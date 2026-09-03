// kmanual.go — RTFM for PROOFS.
//
// The reasoner tier previously diagnosed from recollection. That is the same
// failure mode the sampling guard exists to prevent, one level up: a model
// remembering what a manual says instead of being shown it. On 2026-09-02 the
// reasoner proposed `pragma Loop_Entry` for a missing prefix invariant —
// plausible-sounding, unsupported by the guidance, and it did not land.
//
// So: Kevin selects the manual section DETERMINISTICALLY from the fault class it
// already computes, and the reasoner is instructed to FOLLOW it rather than
// recall it. Selection is machinery; application is the model's job. The prover
// remains the only judge.
//
// Provenance is kept the same way model cards are (see kguard.go): the documents
// live on disk under ~/.ada-factory/manuals with recorded hashes, so "we read the
// manual" is checkable rather than asserted.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// manualExcerpt is one authoritative passage plus where it came from.
type manualExcerpt struct {
	Title  string
	Source string
	Body   string
}

// The four properties are quoted verbatim from AdaCore's SPARK User's Guide,
// "How to Write Loop Invariants". They are the checklist the reasoner must work
// through; naming them is what turns a vague "add an invariant" into a question
// with four decidable parts.
const loopInvariantLaw = `AUTHORITATIVE GUIDANCE — SPARK User's Guide, "How to Write Loop Invariants".
A loop invariant must satisfy ALL FOUR of these properties:

  [INIT]     It should be provable in the first iteration of the loop.
  [INSIDE]   It should allow proving absence of run-time errors and local
             assertions inside the loop.
  [AFTER]    It should allow proving absence of run-time errors, local assertions
             and the subprogram postcondition after the loop.
  [PRESERVE] It should be provable at an arbitrary iteration of the loop,
             assuming that it held at the previous iteration.

PATTERN FOR A POSTCONDITION QUANTIFIED OVER AN ARRAY:
state the invariant with a RESTRICTED quantification range, so that when the
index reaches its last value the restricted range coincides with the full
postcondition. The Guide's own example narrows a postcondition over A'Range to
the loop invariant (for all K in A'First .. Pos => A (K) /= I).

PLACEMENT: the Guide notes it is simpler to relate the invariant to the checks
that follow the loop when the loop invariant is directly followed by the exit
statement. No strict placement rule is mandated.`

// manualFor maps a Kevin fault class to the guidance that governs it. Returning
// no excerpt is a legitimate answer — a fault with no governing passage should
// NOT be given an unrelated one, which would be worse than none.
func manualFor(faultKey string) *manualExcerpt {
	k := strings.ToUpper(faultKey)
	switch {
	case strings.Contains(k, "INVARIANT"),
		strings.Contains(k, "DECOMPOSITION"),
		strings.Contains(k, "RANGE FACT"),
		strings.Contains(k, "POSTCONDITION"):
		return &manualExcerpt{
			Title:  "How to Write Loop Invariants",
			Source: "AdaCore SPARK User's Guide (local copy, hashed)",
			Body:   loopInvariantLaw,
		}
	}
	return nil
}

var explainCodeRE = regexp.MustCompile(`\[(E\d{4})\]`)

// explainCodesIn returns GNATprove's OWN explanation for any [E00nn] codes the
// prover emitted. This is the tool documenting itself: the most authoritative
// text available, and it ships on disk with the toolchain.
func explainCodesIn(proverOutput string) []manualExcerpt {
	seen := map[string]bool{}
	var out []manualExcerpt
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	for _, m := range explainCodeRE.FindAllStringSubmatch(proverOutput, -1) {
		code := m[1]
		if seen[code] {
			continue
		}
		seen[code] = true
		// PORTABLE: the author's cached copy first, then ASK THE INSTALLED TOOL.
		// ⚠ Previously a missing cache file was skipped in silence, so anyone
		// without ~/.ada-factory lost the RTFM rung entirely and their scores
		// were quietly not comparable with ours (2026-09-03). Reading it from
		// the reader's own gnatprove is better than shipping a copy: no
		// redistribution question, and it is THEIR toolchain version.
		source := "GNATprove toolchain, shipped explanation database"
		path := filepath.Join(home, ".ada-factory", "manuals", "gnatprove_explain_codes", code+".md")
		body, err := os.ReadFile(path)
		if err != nil {
			explained, exErr := exec.Command("gnatprove", "--explain="+code).CombinedOutput()
			if exErr != nil || len(strings.TrimSpace(string(explained))) == 0 {
				continue // the toolchain has no entry either; silence beats a wrong page
			}
			body = explained
			source = "gnatprove --explain=" + code + " (local toolchain)"
		}
		out = append(out, manualExcerpt{
			Title:  "GNATprove explain code " + code,
			Source: source,
			Body:   strings.TrimSpace(string(body)),
		})
	}
	return out
}

// BuildManualBrief assembles the guidance for one failing attempt. Empty means
// no governing passage was found — say nothing rather than pad the prompt.
func BuildManualBrief(faultKey, proverOutput string) string {
	var parts []string
	if e := manualFor(faultKey); e != nil {
		parts = append(parts, fmt.Sprintf("### %s\n(%s)\n\n%s", e.Title, e.Source, e.Body))
	}
	for _, e := range explainCodesIn(proverOutput) {
		parts = append(parts, fmt.Sprintf("### %s\n(%s)\n\n%s", e.Title, e.Source, e.Body))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n")
}

// BuildPlanPrompt asks the reasoner for a PLAN before the coder writes anything.
//
// WHY (Tony, 2026-09-02): the reasoner is currently spent on the weakest lever —
// repairing a body the coder has already committed to. Two measurements argue the
// first draw is what matters:
//
//	· the controller repeatedly reports "identical body repeated 2x, 3x" — once the
//	  coder has an approach it does not leave it, so the FIRST generation dominates;
//	· feeding prover output back was measured not to work at all (IR-12, reconfirmed
//	  twice on 2026-09-02), so post-hoc repair is the weak arm by evidence.
//
// This is NOT the July wall restated. That finding
// (fact_forge_qwen_loop_invariant_gap_2026_07_07) is that the coder cannot APPLY an
// inductive invariant handed to it as a fix for an existing body. Conditioning the
// initial generation is a different mechanism and has never been measured. It may
// fail the same way — that is the experiment.
//
// The plan is prose ABOUT the proof, never code: the coder still writes the body, so
// the artifact stays the local model's and the prover still judges it alone.
func BuildPlanPrompt(spec, manual string) string {
	m := ""
	if manual != "" {
		m = "\n\nAUTHORITATIVE MANUAL — follow this, do not rely on recollection:\n" + manual + "\n"
	}
	return "You are planning a SPARK proof BEFORE any code is written." + m +
		"\nBelow is a SPARK specification whose body is still a placeholder. Do NOT write" +
		"\nthe body. Instead state, in at most 8 short lines:" +
		"\n  1. the loop structure the body will need (or 'no loop');" +
		"\n  2. the exact `pragma Loop_Invariant` lines required, written as Ada," +
		"\n     restricted so that at the final index they coincide with the postcondition;" +
		"\n  3. which of [INIT] [INSIDE] [AFTER] [PRESERVE] each invariant is there to satisfy;" +
		"\n  4. any range or overflow fact the prover will need stated explicitly." +
		"\nIf the postcondition is not satisfiable as written, say so in one line and stop." +
		"\n\nSPECIFICATION:\n" + spec
}

// planTimeDecomposition splits the postcondition into its separate theorems for
// the PLAN prompt, before any body exists.
//
// splitConjuncts works from the spec alone — only failingConjunct needs prover
// output — so the reasoner can be told "this Post is N separate obligations"
// before the coder writes anything. Withholding it was asking for a proof plan
// while hiding half the case file.
//
// The synthesiser deliberately is NOT here: synthesisePrefixInvariant needs a
// loop counter, which it reads from a BODY, and at plan time the body is still
// the `pragma Assume (False)` stub. That clue is genuinely post-generation.
func planTimeDecomposition(spec string) string {
	post := extractPost(spec)
	if post == "" {
		return ""
	}
	conjuncts := splitConjuncts(post)
	if len(conjuncts) < 2 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "DECOMPOSITION — this postcondition is %d SEPARATE THEOREMS joined by `and`.\n"+
		"Plan each one individually; an invariant that carries one conjunct out of the loop\n"+
		"may say nothing about another.\n", len(conjuncts))
	for i, c := range conjuncts {
		fmt.Fprintf(&b, "  (%d) %s\n", i+1, truncateConjunct(c))
	}
	return b.String()
}
