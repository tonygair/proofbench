package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// controller.go — THE DIAGNOSER DRIVES THE LOOP.
//
// Until now the retry loop was the model's: generate, fail, append a hint,
// generate again, up to -candidates times. The diagnoser only decorated the
// message. That loop cannot notice it is stuck, and on 2026-09-02 it was
// measured doing exactly that — eight repair attempts on each of three tier-2
// tasks with the judge's own output fed back produced byte-identical judge
// timings every round: the model re-emitted the same body and the loop kept
// asking. Twenty-one of twenty-four attempts were spent re-asking a question
// that had already been answered the same way.
//
// This file inverts the control. The deterministic layer decides what happens
// next; the model is an executor it calls when calling it is worth something.
// Nothing here edits a body — the decision is only whether, and how, to ask
// again. Mechanical repair is a separate increment with its own provenance
// consequences, deliberately not taken here.
//
// The controller is pure: it takes the history of an attempt sequence and
// returns an action. That makes it table-testable, which is the point — a
// deterministic decider you cannot inspect is no better than a model.

// controlAction is what the controller decides to do after an attempt.
type controlAction int

const (
	// actionRetry — ask the model again, with the diagnosis. The default when a
	// class is recognised and the loop is still making progress.
	actionRetry controlAction = iota
	// actionDiversify — the model is repeating itself. Ask again but change the
	// draw (higher temperature / fresh candidate) rather than repeat a question
	// already answered identically.
	actionDiversify
	// actionEscalate — the deterministic hints have not moved it. Hand to the
	// reasoner tier (reasoner.go) for one diagnosis, model still repairs.
	actionEscalate
	// actionHandoff — this model has hit a fixed point the reasoner did not
	// break. Hand the SAME spec to a different local model and let the prover
	// judge. Justified by measurement, not hope: two models at identical
	// 24/26 on tier 1 failed disjoint tasks (2026-09-02).
	actionHandoff
	// actionAbandon — this class is not being won here. Stop spending.
	actionAbandon
)

func (a controlAction) String() string {
	switch a {
	case actionRetry:
		return "retry"
	case actionDiversify:
		return "diversify"
	case actionEscalate:
		return "escalate"
	case actionHandoff:
		return "handoff"
	case actionAbandon:
		return "abandon"
	}
	return "unknown"
}

// attemptOutcome is one pass through generate → compile → prove, as the
// controller sees it. Stage is where it stopped; faultKey is the diagnosis
// class (or "" when unrecognised); bodyHash identifies the artifact produced.
type attemptOutcome struct {
	Stage     string // "compile" | "prove" | "pass"
	FaultKey  string
	BodyHash  string
	Diagnosed bool // a deterministic diagnosis was available for this fault
	Escalated bool // the reasoner tier was spent on this attempt
	HandedOff bool // this attempt was drawn from the OTHER model
}

// controlPolicy bounds what the controller will spend.
type controlPolicy struct {
	MaxAttempts int // hard ceiling on attempts for one spec
	// RepeatsBeforeDiversify — identical (fault, body) pairs tolerated before
	// the controller stops re-asking the same question. 1 means: the moment an
	// attempt repeats its predecessor exactly, change the draw.
	RepeatsBeforeDiversify int
	// StuckBeforeEscalate — how many attempts on the SAME fault class (whether
	// or not the body is identical) before the reasoner tier is worth its cost.
	StuckBeforeEscalate int
	// AllowEscalate — false when no reasoner is configured, in which case the
	// controller goes straight from stuck to abandon rather than escalating
	// into nothing.
	AllowEscalate bool
	// AllowHandoff — a second local coder is configured. The handoff fires only
	// AFTER the reasoner rung has been spent on this fault, so it is the last
	// act before abandon.
	AllowHandoff bool
}

func defaultControlPolicy(allowEscalate bool) controlPolicy {
	return defaultControlPolicyWith(allowEscalate, false)
}

// defaultControlPolicyWith adds the handoff rung when a second coder exists.
func defaultControlPolicyWith(allowEscalate, allowHandoff bool) controlPolicy {
	return controlPolicy{
		MaxAttempts:            8,
		RepeatsBeforeDiversify: 1,
		StuckBeforeEscalate:    3,
		AllowEscalate:          allowEscalate,
		AllowHandoff:           allowHandoff,
	}
}

// bodyFingerprint is the identity the controller uses to tell "the model said
// the same thing again" from "the model tried something new". Whitespace is
// normalised so a reflowed but identical body still counts as a repeat.
func bodyFingerprint(body string) string {
	sum := sha256.Sum256([]byte(strings.Join(strings.Fields(body), " ")))
	return hex.EncodeToString(sum[:8])
}

// faultKeyOf reduces a diagnosis to a stable class key. The diagnosis text
// leads with an upper-case class name ("NESTED-LOOP INVARIANT — TWO PARTS...");
// that name is the key. An unrecognised fault keys on its stage so repeated
// unknown failures are still seen as repetition rather than progress.
func faultKeyOf(diagnosis, stage string) string {
	if diagnosis == "" {
		return stage + ":unclassified"
	}
	for _, line := range strings.Split(diagnosis, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		if i := strings.Index(line, ":"); i > 0 {
			head := strings.TrimSpace(line[:i])
			// class names are shouted; prose sentences are not
			if head == strings.ToUpper(head) && len(head) > 3 {
				return stage + ":" + head
			}
		}
	}
	return stage + ":diagnosed"
}

// decide returns the action to take after `history` (oldest first, the last
// element being the attempt just completed). It never returns actionRetry once
// the budget is spent, and never escalates when escalation is unavailable.
// escalatedOn reports whether the reasoner has already been spent on this fault
// class during this task, so the controller escalates ONCE per class rather than
// paying for the same diagnosis repeatedly.
func escalatedOn(history []attemptOutcome, fault string) bool {
	for _, h := range history {
		if h.Escalated && h.FaultKey == fault {
			return true
		}
	}
	return false
}

// handedOff reports whether the spec has already been passed to the other model.
// The handoff happens at most once: if a second model cannot break the fixed
// point either, that is a finding, not a reason to keep spending.
func handedOff(history []attemptOutcome) bool {
	for _, h := range history {
		if h.HandedOff {
			return true
		}
	}
	return false
}

func decide(history []attemptOutcome, p controlPolicy) (controlAction, string) {
	if len(history) == 0 {
		return actionRetry, "no history yet"
	}
	last := history[len(history)-1]
	if last.Stage == "pass" {
		return actionAbandon, "already passed; nothing to decide"
	}
	budgetSpent := len(history) >= p.MaxAttempts

	// Repetition: the model has produced the same artifact for the same fault.
	// Re-asking has been measured not to work, so change the draw instead.
	repeats := 0
	for i := len(history) - 2; i >= 0; i-- {
		if history[i].BodyHash == last.BodyHash && history[i].FaultKey == last.FaultKey {
			repeats++
			continue
		}
		break
	}
	if repeats >= p.RepeatsBeforeDiversify {
		// Diversifying has itself already been tried this many times; the draw
		// is not the problem.
		if repeats >= p.StuckBeforeEscalate {
			// Escalate ONCE per fault class. Re-buying the same diagnosis for a
			// body that has not changed is spending without a question changing;
			// the next rung is a different model, not a repeat purchase.
			if p.AllowEscalate && !escalatedOn(history, last.FaultKey) {
				return actionEscalate, fmt.Sprintf("identical body for %s across %d attempts", last.FaultKey, repeats+1)
			}
			if p.AllowHandoff && !handedOff(history) {
				return actionHandoff, fmt.Sprintf("identical body for %s across %d attempts; reasoner spent — hand to the other model", last.FaultKey, repeats+1)
			}
			if p.AllowEscalate {
				return actionAbandon, fmt.Sprintf("identical body for %s across %d attempts; reasoner and handoff both spent", last.FaultKey, repeats+1)
			}
			return actionAbandon, fmt.Sprintf("identical body for %s across %d attempts and no reasoner configured", last.FaultKey, repeats+1)
		}
		return actionDiversify, fmt.Sprintf("identical body repeated %d× on %s", repeats+1, last.FaultKey)
	}

	// Stuck on a class: different bodies, same fault, several attempts running.
	// The deterministic hint is not landing; a stronger diagnosis might.
	sameFault := 1
	for i := len(history) - 2; i >= 0; i-- {
		if history[i].FaultKey == last.FaultKey {
			sameFault++
			continue
		}
		break
	}
	if sameFault >= p.StuckBeforeEscalate {
		if p.AllowEscalate && !escalatedOn(history, last.FaultKey) {
			return actionEscalate, fmt.Sprintf("%d consecutive attempts on %s", sameFault, last.FaultKey)
		}
		// The reasoner has already been spent on this fault and it still has not
		// moved. A DIFFERENT model is the next deterministic rung.
		if p.AllowHandoff && !handedOff(history) {
			return actionHandoff, fmt.Sprintf("%d consecutive attempts on %s; reasoner spent — hand to the other model", sameFault, last.FaultKey)
		}
		if p.AllowEscalate {
			return actionAbandon, fmt.Sprintf("%d consecutive attempts on %s; reasoner and handoff both spent", sameFault, last.FaultKey)
		}
		return actionAbandon, fmt.Sprintf("%d consecutive attempts on %s and no reasoner configured", sameFault, last.FaultKey)
	}

	// The budget is spent and no escalation rung claimed this attempt.
	if budgetSpent {
		return actionAbandon, fmt.Sprintf("attempt budget spent (%d)", p.MaxAttempts)
	}

	// An unrecognised fault with no diagnosis to offer is a poor bet: the model
	// gets a raw dump, which IR-12 says it cannot act on. Escalate sooner.
	if !last.Diagnosed && sameFault >= 2 {
		if p.AllowEscalate {
			return actionEscalate, "undiagnosed fault repeated; raw output is not actionable (IR-12)"
		}
		return actionAbandon, "undiagnosed fault repeated and no reasoner configured"
	}

	return actionRetry, "diagnosed and still moving"
}
