// kguard.go — the sampling guard.
//
// WHY THIS EXISTS (2026-09-02): Rosie (qwen3.8-27b-ada:v0.3) was benchmarked at
// temperature 0.2 with thinking left ON and num_predict 16384. Her own model card
// FORBIDS near-greedy decoding ("performance degradation and endless repetitions")
// and the qwen3.x line has thinking ON by default, so chain-of-thought ate the
// token budget before any Ada was emitted. The resulting "Rosie fails to compile"
// was a HARNESS verdict wearing a MODEL's name.
//
// The settings were already known — read, sourced and written up the same day —
// but they lived in prose and in ONE harness's code (oneshot-bench c144fb5).
// wufree never inherited them. So the fix is not "set the values here too": it is
// to make the registry the single source of truth and REFUSE to run outside it.
//
// The guard NEVER silently corrects. A silent correction hides a misconfiguration;
// a refusal names it. Fail-closed: an unregistered model is refused, because an
// unknown sampling policy is exactly the condition that produced the bad run.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SamplingPolicy is the card-authoritative decoding configuration for one model.
// Zero values are meaningful, so presence is tracked by the registry entry itself.
// Provenance is the PROOF that a sampling policy was read from the manual
// rather than invented. It records where the card was fetched from, a hash of
// the exact bytes read, and the clauses relied upon. Because the hash is of a
// re-fetchable document, the claim "we RTFM'd" is CHECKABLE by anyone later:
// re-fetch, re-hash, compare. A mismatch does not mean someone lied — it means
// the card MOVED and must be re-read before the policy is trusted again.
type Provenance struct {
	SourceURL   string `json:"source_url"`
	DocSHA256   string `json:"doc_sha256"`
	DocBytes    int    `json:"doc_bytes"`
	RetrievedAt string `json:"retrieved_utc"`
	QuotedNon   string `json:"quoted_non_thinking"`
	QuotedThink string `json:"quoted_thinking"`
	QuotedBan   string `json:"quoted_prohibition"`
	QuotedLen   string `json:"quoted_output_length"`
	// SourceLangScan is order (2) of the standing pair: the source-language
	// community scan. The card is the baseline; the forums carry hardware- and
	// version-specific colour the card omits. A policy evidenced by the card
	// ALONE satisfies only half the order.
	SourceLangScan *struct {
		ScannedAt   string   `json:"scanned_utc"`
		Communities string   `json:"communities"`
		Brief       string   `json:"brief"`
		Findings    []string `json:"findings_the_card_lacks"`
	} `json:"_source_language_scan"`
}

// Attest reports why the policy should be believed, for the run log.
func (p *Provenance) Attest() string {
	base := fmt.Sprintf("card %s (%d bytes, sha256 %s…, read %s)",
		p.SourceURL, p.DocBytes, p.DocSHA256[:12], p.RetrievedAt)
	if p.SourceLangScan == nil {
		return base + " · ⚠ NO source-language scan (half the standing order)"
	}
	return base + fmt.Sprintf(" · source-lang scan %s (%d findings the card lacks)",
		p.SourceLangScan.ScannedAt, len(p.SourceLangScan.Findings))
}

type SamplingPolicy struct {
	Family            string      `json:"family"`
	Think             bool        `json:"think"`
	Temperature       float64     `json:"temperature"`
	TopP              float64     `json:"top_p"`
	TopK              int         `json:"top_k"`
	MinP              float64     `json:"min_p"`
	PresencePenalty   float64     `json:"presence_penalty"`
	RepetitionPenalty float64     `json:"repetition_penalty"`
	NumPredict        int         `json:"num_predict"`
	MinTemperature    float64     `json:"min_temperature"`
	Authority         string      `json:"_authority"`
	Provenance        *Provenance `json:"_provenance"`
}

type registryEntry struct {
	Provider string          `json:"provider"`
	APIModel string          `json:"api_model"`
	Sampling *SamplingPolicy `json:"sampling"`
}

// GuardError reports a refusal to run. It names the violated clause and the
// authority for it, so the operator is told what to fix rather than that
// something is merely wrong.
type GuardError struct {
	Model  string
	Clause string
	Detail string
}

func (e *GuardError) Error() string {
	return fmt.Sprintf("SAMPLING GUARD REFUSED %s — %s: %s", e.Model, e.Clause, e.Detail)
}

// loadRegistry reads the shared model registry. The path is the same file
// oneshot-bench uses, deliberately: two harnesses, one source of truth.
func loadRegistry(path string) (map[string]registryEntry, error) {
	if path == "" {
		// PORTABLE: a fresh clone must run with no setup and without writing to
		// anyone's home directory.
		//
		// ⚠ The earlier version checked only the CURRENT DIRECTORY, while the
		// README tells the reader to run from the package root with the binary
		// at harness/proofbench — so the bundled registry beside the binary was
		// never found, the lookup fell through to the author's home, and every
		// task REFUSED for anyone but the author (2026-09-03). The guard was
		// working exactly as designed; the packaging was wrong.
		candidates := []string{"models.json"}
		if exe, err := os.Executable(); err == nil {
			dir := filepath.Dir(exe)
			candidates = append(candidates,
				filepath.Join(dir, "models.json"),       // beside the binary
				filepath.Join(dir, "..", "models.json"), // package root
			)
		}
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, ".ada-factory", "models.json"))
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				path = c
				break
			}
		}
		if path == "" {
			return nil, fmt.Errorf("no model registry found — looked in: %s", strings.Join(candidates, ", "))
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read model registry %s: %w", path, err)
	}
	// The registry carries documentation keys (e.g. "_comment") alongside model
	// entries, so decode loosely and skip anything that is not an object. A
	// malformed ENTRY is still an error — only non-entries are tolerated.
	var loose map[string]json.RawMessage
	if err := json.Unmarshal(raw, &loose); err != nil {
		return nil, fmt.Errorf("parse model registry %s: %w", path, err)
	}
	reg := make(map[string]registryEntry, len(loose))
	for key, raw := range loose {
		if len(raw) == 0 || raw[0] != '{' {
			continue // documentation key, not a model entry
		}
		var e registryEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("parse model registry %s: entry %q: %w", path, key, err)
		}
		reg[key] = e
	}
	return reg, nil
}

// PolicyFor resolves the sampling policy for an ollama tag (e.g.
// "qwen3.8-27b-ada:v0.3") by matching the registry's api_model field.
// A model with no entry, or an entry with no sampling block, is REFUSED:
// fail-closed is the whole point of the guard.
func PolicyFor(reg map[string]registryEntry, tag string) (*SamplingPolicy, error) {
	for key, e := range reg {
		if e.APIModel != tag {
			continue
		}
		if e.Sampling != nil && e.Sampling.Provenance == nil {
			return nil, &GuardError{Model: tag, Clause: "unproven settings",
				Detail: fmt.Sprintf("registry key %q declares a sampling policy with NO \"_provenance\" block — no evidence the model card was ever read. RTFM, then record source_url + doc_sha256 + the quoted clauses", key)}
		}
		if e.Sampling == nil {
			return nil, &GuardError{Model: tag, Clause: "no sampling policy",
				Detail: fmt.Sprintf("registry key %q has no \"sampling\" block; add one (RTFM the model card first) before benchmarking", key)}
		}
		return e.Sampling, nil
	}
	return nil, &GuardError{Model: tag, Clause: "unregistered model",
		Detail: "no registry entry has api_model=" + tag + "; add an entry with a card-authoritative \"sampling\" block"}
}

// Check verifies a proposed decoding configuration against the policy and
// refuses on violation. It reports EVERY violated clause, not just the first —
// a run refused twice for one reason each is two wasted cycles.
func (p *SamplingPolicy) Check(tag string, temp float64, numPredict int, think bool) error {
	var bad []string
	if temp < p.MinTemperature {
		bad = append(bad, fmt.Sprintf("temperature %.2f is below the floor %.2f (card forbids greedy/near-greedy: endless repetitions)", temp, p.MinTemperature))
	}
	if numPredict < p.NumPredict {
		bad = append(bad, fmt.Sprintf("num_predict %d is below the card's %d (thinking or long bodies truncate before the answer)", numPredict, p.NumPredict))
	}
	if think != p.Think {
		bad = append(bad, fmt.Sprintf("think=%v but policy requires think=%v", think, p.Think))
	}
	if len(bad) == 0 {
		return nil
	}
	return &GuardError{Model: tag, Clause: "config violates the model card",
		Detail: strings.Join(bad, "; ") + " [authority: " + p.Authority + "]"}
}
