// wufree — the UNFETTERED loop.
//
// wurun.go honours the benchmark's protocol: the model gets a fixed prompt, one
// shot per round, and NEVER sees GNATprove's output — only the seat does, and
// the seat rewrites prose between rounds. That restriction is what makes a Wu
// score comparable to a benchmark column.
//
// This driver deliberately lifts it. The prover's exact complaint, plus the
// model's own previous sections, go straight back to the model and it repairs
// its own body. It iterates until the harness says pass or the attempt budget
// runs out. Optionally it draws several candidates per attempt and keeps the
// first that proves.
//
// WHAT THIS IS AND IS NOT: results from this driver are ENGINEERING evidence.
// They are NOT benchmark-comparable and must never be quoted in a leaderboard
// column beside a fixed-prompt score. Same judge, same frozen spec, same task
// files — but a closed loop and unbounded attempts.
//
// The spec, preamble, Pre and Post are still frozen: this driver only ever
// rewrites <vc-helpers> and <vc-code>, exactly as wurun.go does. The theorem is
// never weakened.
//
// Usage:
//
//	go run wufree.go -track DIR -run DIR -model M [-only T1,T2] [-attempts N] [-candidates N]
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	sectionCache = map[string]*regexp.Regexp{}
	fenceRE      = regexp.MustCompile("(?s)```(?:json)?\\s*(\\[.*?\\])\\s*```")
)

func sectionRE(tag string) *regexp.Regexp {
	if re, ok := sectionCache[tag]; ok {
		return re
	}
	re := regexp.MustCompile(`(?s)(--\s*<` + tag + `>\n)(.*?)([ \t]*--\s*</` + tag + `>)`)
	sectionCache[tag] = re
	return re
}

func replaceSection(text, tag, body string) (string, error) {
	re := sectionRE(tag)
	if !re.MatchString(text) {
		return "", fmt.Errorf("task file has no <%s> section", tag)
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return re.ReplaceAllStringFunc(text, func(m string) string {
		parts := re.FindStringSubmatch(m)
		return parts[1] + body + parts[3]
	}), nil
}

func sectionOf(text, tag string) string {
	re := sectionRE(tag)
	if m := re.FindStringSubmatch(text); m != nil {
		return m[2]
	}
	return ""
}

// extractJSONArray tolerates thinking models: skip past a final </think>, prefer
// the LAST array (the answer) over the first (quoted while reasoning).
func extractJSONArray(text string) ([]string, error) {
	var candidates []string
	if i := strings.LastIndex(text, "</think>"); i != -1 {
		text = text[i+len("</think>"):]
	}
	if m := fenceRE.FindStringSubmatch(text); m != nil {
		candidates = append(candidates, m[1])
	}
	if ls := strings.LastIndex(text, "["); ls != -1 {
		if e := strings.LastIndex(text, "]"); e > ls {
			candidates = append(candidates, text[ls:e+1])
		}
	}
	if s, e := strings.Index(text, "["), strings.LastIndex(text, "]"); s != -1 && e > s {
		candidates = append(candidates, text[s:e+1])
	}
	for _, c := range candidates {
		var v []string
		if err := json.Unmarshal([]byte(c), &v); err == nil && len(v) == 2 {
			return v, nil
		}
	}
	return nil, errors.New("no JSON array of two strings in the response")
}

type genOptions struct {
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"top_p,omitempty"`
	TopK            int     `json:"top_k,omitempty"`
	MinP            float64 `json:"min_p"`
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	RepeatPenalty   float64 `json:"repeat_penalty,omitempty"`
	NumPredict      int     `json:"num_predict"`
	NumCtx          int     `json:"num_ctx"`
}

type genRequest struct {
	Model   string     `json:"model"`
	Prompt  string     `json:"prompt"`
	Stream  bool       `json:"stream"`
	Think   bool       `json:"think"`
	Options genOptions `json:"options"`
}

type genResponse struct {
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	EvalCount int    `json:"eval_count"`
}

// generate emits one candidate. The decoding configuration comes from the
// guarded SamplingPolicy (see kguard.go) — never from a literal here, which is
// how wufree came to run Rosie in a mode her own model card forbids.
func generate(ctx context.Context, rail, model, prompt string, temp float64, pol *SamplingPolicy, timeout time.Duration) (genResponse, float64, error) {
	body, err := json.Marshal(genRequest{
		Model: model, Prompt: prompt, Stream: false,
		Think: pol.Think,
		Options: genOptions{
			Temperature: temp, TopP: pol.TopP, TopK: pol.TopK, MinP: pol.MinP,
			PresencePenalty: pol.PresencePenalty, RepeatPenalty: pol.RepetitionPenalty,
			NumPredict: pol.NumPredict, NumCtx: 32768,
		},
	})
	if err != nil {
		return genResponse{}, 0, err
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout+30*time.Second)
	defer cancel()
	started := time.Now()
	cmd := exec.CommandContext(callCtx, "curl", "-s", "--max-time",
		fmt.Sprintf("%d", int(timeout.Seconds())),
		rail+"/api/generate", "-H", "Content-Type: application/json", "-d", string(body))
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return genResponse{}, time.Since(started).Seconds(), fmt.Errorf("inference call: %w (%s)", err, truncate(errb.String(), 200))
	}
	var r genResponse
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		return genResponse{}, time.Since(started).Seconds(), fmt.Errorf("decode: %w", err)
	}
	return r, time.Since(started).Seconds(), nil
}

func askHarness(ctx context.Context, trackDir, file string, timeout time.Duration) (json.RawMessage, float64, error) {
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	started := time.Now()
	cmd := exec.CommandContext(callCtx, "python3",
		filepath.Join(trackDir, "judge", "spark_verify.py"), "--json", file)
	cmd.Dir = trackDir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	_ = cmd.Run()
	line := strings.TrimSpace(out.String())
	if line == "" {
		return nil, time.Since(started).Seconds(), fmt.Errorf("harness produced no verdict: %s", truncate(errb.String(), 300))
	}
	return json.RawMessage(line), time.Since(started).Seconds(), nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

type verdict struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type attemptRecord struct {
	Task        string          `json:"task"`
	Attempt     int             `json:"attempt"`
	Candidate   int             `json:"candidate"`
	Model       string          `json:"model"`
	Temperature float64         `json:"temperature"`
	Status      string          `json:"status"`
	Harness     json.RawMessage `json:"harness,omitempty"`
	Detail      string          `json:"detail,omitempty"`
	InferSec    float64         `json:"infer_s"`
	JudgeSec    float64         `json:"judge_s"`
	EvalCount   int             `json:"eval_count"`
	AtUTC       string          `json:"at_utc"`
	FedBack     bool            `json:"fed_prover_output"`

	// --- Kevin's reasoning, recorded rather than merely printed ---------------
	// Until 2026-09-02 these existed only on stdout, which meant the determinism
	// could be narrated but not measured. A deterministic decider you cannot
	// inspect is no better than a model.
	Stage         string `json:"stage,omitempty"`     // compile | prove | pass
	FaultKey      string `json:"fault_key,omitempty"` // the diagnosis class
	BodyHash      string `json:"body_hash,omitempty"` // fixed-point fingerprint
	Diagnosed     bool   `json:"diagnosed"`           // a deterministic diagnosis existed
	Decision      string `json:"decision,omitempty"`  // retry|diversify|escalate|abandon
	DecisionWhy   string `json:"decision_why,omitempty"`
	Escalated     bool   `json:"escalated"`   // the reasoner tier was spent
	HandedOff     bool   `json:"handed_off"`  // drawn from the OTHER coder
	ManualUsed    bool   `json:"manual_used"` // authoritative guidance was attached
	ReasonerModel string `json:"reasoner_model,omitempty"`
	ReasonerFix   string `json:"reasoner_fix,omitempty"`
	ReasonerErr   string `json:"reasoner_error,omitempty"`
	// Sampling actually used, so a run carries its own guard evidence.
	Think      bool    `json:"think"`
	TopP       float64 `json:"top_p,omitempty"`
	TopK       int     `json:"top_k,omitempty"`
	NumPredict int     `json:"num_predict,omitempty"`
}

// buildDiagnosePrompt mirrors ada-factory's reasoner tier: a STRONG model is
// asked for ONE structural diagnosis, never for a body. The local coder still
// writes the code, so the artifact stays the local model's and the proof gate
// judges it regardless of who hinted.
// buildDiagnosePrompt now carries the MANUAL. Kevin selects the governing
// passage from the fault class it already computed; the reasoner is told to
// FOLLOW that text, not to recall it. Where no passage governs the fault, the
// brief is empty and the prompt is unchanged — an unrelated page is worse than
// none.
func buildDiagnosePrompt(spec, body, proveErrs, faultKey string) string {
	manual := BuildManualBrief(faultKey, proveErrs)
	if manual != "" {
		manual = "\n\nAUTHORITATIVE MANUAL — follow this text, do not rely on recollection.\n" +
			"Work through the named properties explicitly and say which one the current body\n" +
			"violates before giving the fix.\n\n" + manual + "\n"
	}
	return manual +
		"A SPARK package body was REJECTED by `gnatprove --level=2` (or failed to compile).\n" +
		"Diagnose the SINGLE most important structural fix the author must make — the exact\n" +
		"`pragma Loop_Invariant` to add (write the Ada), a missing boundary guard, a `'Loop_Entry`\n" +
		"frame invariant, a loop-structure change, etc. Be concrete and specific to THIS body.\n" +
		"Do NOT rewrite the whole body — the author applies your diagnosis. Answer in at most 4\n" +
		"sentences, starting with 'FIX:'.\n\n" +
		"SPEC + TASK FILE:\n" + spec + "\n\nSECTIONS THAT FAILED:\n" + body + "\n\nthe judge said:\n" + proveErrs
}

// reasonerDiagnose asks the sovereign reasoner (a local ollama model) for one
// hint. Errors are non-fatal: the deterministic diagnosis still stands.
func reasonerDiagnose(ctx context.Context, rail, model, spec, body, proveErrs, faultKey string, pol *SamplingPolicy, timeout time.Duration) (string, error) {
	if model == "" {
		return "", nil
	}
	resp, _, err := generate(ctx, rail, model, buildDiagnosePrompt(spec, body, proveErrs, faultKey), pol.Temperature, pol, timeout)
	if err != nil {
		return "", err
	}
	out := resp.Response
	if i := strings.LastIndex(out, "</think>"); i != -1 {
		out = out[i+len("</think>"):]
	}
	return strings.TrimSpace(out), nil
}

const repairHeader = `The Ada body you wrote for this SPARK task was REJECTED. Below is the exact
output of the judge (GNATprove at --level=2), then the two sections you wrote.

Read the complaint literally. It names a file, a line, and the check it could
not prove. Fix THAT. Do not restructure code that the judge did not complain
about. Do not weaken or restate the postcondition - the specification is fixed
and you may not edit it.

=== JUDGE OUTPUT (this is what you must satisfy) ===
%s
=== YOUR PREVIOUS <vc-helpers> ===
%s
=== YOUR PREVIOUS <vc-code> ===
%s
=== END ===

Now emit the CORRECTED sections, in the same output format as before: ONLY a
JSON array of exactly 2 strings, element 1 = <vc-helpers>, element 2 =
<vc-code>. First character [ and last character ].
`

func main() {
	var (
		trackDir     = flag.String("track", "", "path to the SPARK track tree")
		runDir       = flag.String("run", "", "run directory (holds prose/, results)")
		model        = flag.String("model", "qwen3-coder:30b", "ollama model")
		rail         = flag.String("rail", "http://127.0.0.1:11434", "ollama rail")
		only         = flag.String("only", "", "comma-separated task stems")
		proseRound   = flag.String("prose", "D", "which round-<X>.txt to use as the standing prose")
		attempts     = flag.Int("attempts", 8, "max repair attempts per task")
		candidates   = flag.Int("candidates", 1, "candidates drawn per attempt (first that proves wins)")
		temperature  = flag.Float64("temperature", 0, "decoding temperature; 0 = use the model card's value from the registry (recommended). Any value below the card's floor is REFUSED by the sampling guard.")
		planFile     = flag.String("plan-file", "", "EXPERIMENT: hand the coder a FIXED plan from this file instead of asking a reasoner. Used to test whether the model can act on a KNOWN-GOOD invariant — separating a reasoning gap from a prover-automation gap.")
		reasonFirst  = flag.Bool("reason-first", false, "ask the reasoner for a PLAN before the coder generates, instead of only escalating after failure")
		reasonerRail = flag.String("reasoner-rail", "", "rail for the REASONER (default: same as -rail). Lets a big reasoner live on another host while the coder stays local.")
		handoffM     = flag.String("handoff", "", "SECOND local coder. When the reasoner has been spent on a fault and the model is still at a fixed point, the same spec is handed to this model; the prover still judges. Measured worth: two 24/26 models failed disjoint tier-1 tasks (2026-09-02). Empty = no handoff rung.")
		reasonerM    = flag.String("reasoner", "", "sovereign REASONER model (local ollama tag, e.g. deepseek-r1:32b). When set, the controller escalates to it for one structural diagnosis instead of abandoning. Empty = no reasoner tier.")
		timeout      = flag.Duration("timeout", 20*time.Minute, "per-call timeout")
	)
	flag.Parse()
	if *trackDir == "" || *runDir == "" {
		fmt.Fprintln(os.Stderr, "wufree: -track and -run are required")
		os.Exit(2)
	}
	if err := run(*trackDir, *runDir, *model, *rail, *only, *proseRound, *attempts, *candidates, *reasonerM, *reasonerRail, *handoffM, *reasonFirst, *planFile, *temperature, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, "wufree:", err)
		os.Exit(1)
	}
}

func run(trackDir, runDir, model, rail, only, proseRound string, attempts, candidates int, reasoner, reasonerRail, handoff string, reasonFirst bool, planFile string, temp float64, timeout time.Duration) error {
	// THE SAMPLING GUARD — fail-closed, before any tokens are spent. A model
	// benchmarked outside its card's decoding envelope produces a verdict about
	// the HARNESS wearing the MODEL's name (Rosie, 2026-09-02).
	reg, err := loadRegistry("")
	if err != nil {
		return fmt.Errorf("sampling guard: %w", err)
	}
	pol, err := PolicyFor(reg, model)
	if err != nil {
		return err
	}
	effTemp := temp
	if effTemp == 0 {
		effTemp = pol.Temperature // 0 means "use the card"
	}
	if err := pol.Check(model, effTemp, pol.NumPredict, pol.Think); err != nil {
		return err
	}
	fmt.Printf("sampling guard: %s OK — think=%v temp=%.2f top_p=%.2f top_k=%d num_predict=%d presence=%.2f\n",
		model, pol.Think, effTemp, pol.TopP, pol.TopK, pol.NumPredict, pol.PresencePenalty)
	fmt.Printf("  RTFM proof:   %s\n", pol.Provenance.Attest())

	baseTemp := effTemp

	reasonerPol := pol
	if reasoner != "" {
		reasonerPol, err = PolicyFor(reg, reasoner)
		if err != nil {
			return err
		}
		fmt.Printf("sampling guard: %s (reasoner) OK — think=%v\n", reasoner, reasonerPol.Think)
	}

	if reasonerRail == "" {
		reasonerRail = rail
	}
	handoffPol := pol
	if handoff != "" {
		handoffPol, err = PolicyFor(reg, handoff)
		if err != nil {
			return err
		}
		if err := handoffPol.Check(handoff, handoffPol.Temperature, handoffPol.NumPredict, handoffPol.Think); err != nil {
			return err
		}
		fmt.Printf("sampling guard: %s (handoff) OK — think=%v temp=%.2f num_predict=%d\n",
			handoff, handoffPol.Think, handoffPol.Temperature, handoffPol.NumPredict)
	}

	proseBytes, err := os.ReadFile(filepath.Join(runDir, "prose", "round-"+proseRound+".txt"))
	if err != nil {
		return fmt.Errorf("read prose: %w", err)
	}
	prose := string(proseBytes)

	files, err := filepath.Glob(filepath.Join(trackDir, "tasks", "*.ada"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	wanted := map[string]bool{}
	for _, t := range strings.Split(only, ",") {
		if t = strings.TrimSpace(t); t != "" {
			wanted[t] = true
		}
	}
	var tasks []string
	for _, f := range files {
		stem := strings.TrimSuffix(filepath.Base(f), ".ada")
		if len(wanted) == 0 || wanted[stem] {
			tasks = append(tasks, f)
		}
	}
	// ⚠ A misspelt -only used to select nothing and say only "no task files
	// selected", which reads as "the track is empty" rather than "you asked for
	// a task that does not exist". Name the names: a silent nothing is the
	// failure mode this whole benchmark exists to complain about.
	if len(wanted) > 0 {
		var unmatched []string
		for w := range wanted {
			found := false
			for _, f := range files {
				if strings.TrimSuffix(filepath.Base(f), ".ada") == w {
					found = true
					break
				}
			}
			if !found {
				unmatched = append(unmatched, w)
			}
		}
		if len(unmatched) > 0 {
			sort.Strings(unmatched)
			return fmt.Errorf("no task matches %s in %s — task stems carry the -spec suffix (e.g. NpAdd-spec, not NpAdd); %d task files are present",
				strings.Join(unmatched, ", "), filepath.Join(trackDir, "tasks"), len(files))
		}
	}
	if len(tasks) == 0 {
		return errors.New("no task files selected")
	}

	outDir := filepath.Join(runDir, "unfettered")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	sink, err := os.Create(filepath.Join(runDir, "results-unfettered.jsonl"))
	if err != nil {
		return err
	}
	defer sink.Close()
	enc := json.NewEncoder(sink)

	ctx := context.Background()
	for _, taskPath := range tasks {
		stem := strings.TrimSuffix(filepath.Base(taskPath), ".ada")
		original, err := os.ReadFile(taskPath)
		if err != nil {
			return err
		}
		prompt := strings.ReplaceAll(prose, "{{TASK_FILE}}", string(original))
		prevHelpers, prevCode, lastJudge := "", "", ""
		passed := false

		// Kevin in the loop: the raw judge dump is replaced by a NAMED fault class
		// where one is recognised (IR-12 — the model repairs but cannot diagnose),
		// and the controller decides when re-asking has stopped being worth it.
		var history []attemptOutcome
		policy := defaultControlPolicyWith(reasoner != "", handoff != "") // escalation only if a reasoner exists
		policy.MaxAttempts = attempts
		lastDiag := ""

		// REASON-FIRST: ask the reasoner for a PLAN before the coder writes anything,
		// rather than spending it repairing a body already committed to.
		plan := ""
		if planFile != "" {
			if pb, perr := os.ReadFile(planFile); perr == nil {
				plan = strings.TrimSpace(string(pb))
				fmt.Printf("  %-18s plan:  FIXED plan from %s (known-good, not model-derived)\n", stem, planFile)
			}
		}
		if reasonFirst && reasoner != "" {
			pr, _, perr := generate(ctx, reasonerRail, reasoner,
				BuildPlanPrompt(string(original),
					BuildManualBrief("prove:INVARIANT", "")+"\n\n"+planTimeDecomposition(string(original))),
				reasonerPol.Temperature, reasonerPol, timeout)
			if perr == nil {
				plan = strings.TrimSpace(pr.Response)
				if i := strings.LastIndex(plan, "</think>"); i != -1 {
					plan = strings.TrimSpace(plan[i+len("</think>"):])
				}
				if plan != "" {
					fmt.Printf("  %-18s plan:  %s\n", stem, truncate(plan, 150))
				}
			} else {
				fmt.Printf("  %-18s plan:  reasoner failed — %v\n", stem, perr)
			}
		}
		curModel, curPol, viaHandoff := model, pol, false
		handoffFunded, escalateFunded := false, false
		attempts := attempts // per-task budget; a handoff may extend it once
		for attempt := 1; attempt <= attempts && !passed; attempt++ {
			ask := prompt
			if plan != "" {
				ask = prompt + "\n\nPROOF PLAN (from a reasoning model, follow it):\n" + plan + "\n"
			}
			fedBack := false
			if attempt > 1 {
				feedback := lastJudge
				if lastDiag != "" {
					feedback = lastDiag + "\n\nRaw judge output:\n" + lastJudge
				}
				ask = prompt + "\n\n" + fmt.Sprintf(repairHeader, feedback, prevHelpers, prevCode)
				fedBack = true
			}
			for cand := 1; cand <= candidates && !passed; cand++ {
				// Candidate 1 sits at the card's temperature; later candidates
				// diversify UPWARD from it. Never downward — below the card's
				// floor is the repetition trap the guard exists to prevent.
				temp := baseTemp
				if cand > 1 {
					temp = baseTemp + 0.2*float64(cand-1)
				}
				resp, inferSec, genErr := generate(ctx, rail, curModel, ask, temp, curPol, timeout)
				rec := attemptRecord{
					Task: stem, Attempt: attempt, Candidate: cand, Model: curModel,
					Temperature: temp, InferSec: inferSec, EvalCount: resp.EvalCount,
					AtUTC: time.Now().UTC().Format(time.RFC3339), FedBack: fedBack,
					Think: curPol.Think, TopP: curPol.TopP, TopK: curPol.TopK, NumPredict: curPol.NumPredict,
					HandedOff: viaHandoff,
				}
				if genErr != nil {
					rec.Status, rec.Detail = "inference_error", genErr.Error()
					_ = enc.Encode(rec)
					fmt.Printf("  %-18s a%d c%d  %-16s infer=%5.1fs\n", stem, attempt, cand, rec.Status, inferSec)
					continue
				}
				parts, perr := extractJSONArray(resp.Response)
				if perr != nil {
					rec.Status, rec.Detail = "unparseable", perr.Error()
					_ = enc.Encode(rec)
					fmt.Printf("  %-18s a%d c%d  %-16s infer=%5.1fs\n", stem, attempt, cand, rec.Status, inferSec)
					continue
				}
				withHelpers, err := replaceSection(string(original), "vc-helpers", parts[0])
				if err != nil {
					return err
				}
				full, err := replaceSection(withHelpers, "vc-code", parts[1])
				if err != nil {
					return err
				}
				genPath := filepath.Join(outDir, stem+".ada")
				if err := os.WriteFile(genPath, []byte(full), 0o644); err != nil {
					return err
				}
				abs, _ := filepath.Abs(genPath)
				harnessLine, judgeSec, herr := askHarness(ctx, trackDir, abs, timeout)
				rec.JudgeSec = judgeSec
				if herr != nil {
					rec.Status, rec.Detail = "harness_error", herr.Error()
					_ = enc.Encode(rec)
					fmt.Printf("  %-18s a%d c%d  %-16s infer=%5.1fs judge=%5.1fs\n", stem, attempt, cand, rec.Status, inferSec, judgeSec)
					continue
				}
				var v verdict
				_ = json.Unmarshal(harnessLine, &v)
				rec.Harness = harnessLine
				rec.Status = "harness:" + v.Status
				fmt.Printf("  %-18s a%d c%d  %-16s infer=%5.1fs judge=%5.1fs\n", stem, attempt, cand, rec.Status, inferSec, judgeSec)

				prevHelpers, prevCode = parts[0], parts[1]
				lastJudge = truncate(v.Detail, 3000)
				if lastJudge == "" {
					lastJudge = truncate(string(harnessLine), 3000)
				}
				if v.Status == "pass" {
					passed = true
					rec.Stage = "pass"
					_ = enc.Encode(rec)
					// keep the winning file under its own name
					_ = os.WriteFile(filepath.Join(outDir, stem+".PROVED.ada"), []byte(full), 0o644)
					break
				}

				// Kevin classifies the failure; the controller decides what next.
				stage := "prove"
				if v.Status == "compile_error" {
					stage = "compile"
					lastDiag = diagnoseCompileFailure(v.Detail, parts[1])
				} else {
					lastDiag = diagnoseProofFailure(v.Detail, string(original), parts[1])
				}
				outcome := attemptOutcome{
					Stage:     stage,
					FaultKey:  faultKeyOf(lastDiag, stage),
					BodyHash:  bodyFingerprint(parts[1]),
					Diagnosed: lastDiag != "",
					HandedOff: viaHandoff,
				}
				history = append(history, outcome)
				action, why := decide(history, policy)
				rec.Stage, rec.FaultKey = outcome.Stage, outcome.FaultKey
				rec.BodyHash, rec.Diagnosed = outcome.BodyHash, outcome.Diagnosed
				rec.Decision, rec.DecisionWhy = action.String(), why
				fmt.Printf("      controller: %-9s — %s%s\n", action, why,
					map[bool]string{true: "", false: "  [UNDIAGNOSED]"}[lastDiag != ""])
				switch action {
				case actionEscalate:
					if attempt >= attempts && !escalateFunded {
						attempts++
						escalateFunded = true
						fmt.Printf("      budget:     +1 attempt so the diagnosis can be applied\n")
					}
					// The deterministic hints have stopped moving it. Spend the
					// reasoner: ONE structural diagnosis, the local coder still
					// writes the body.
					rdiag, rerr := reasonerDiagnose(ctx, reasonerRail, reasoner,
						string(original), parts[0]+"\n"+parts[1], v.Detail, outcome.FaultKey, reasonerPol, timeout)
					rec.Escalated, rec.ReasonerModel = true, reasoner
					rec.ManualUsed = BuildManualBrief(outcome.FaultKey, v.Detail) != ""
					if rec.ManualUsed {
						fmt.Printf("      manual:     attached for %s (RTFM, not recall)\n", outcome.FaultKey)
					}
					history[len(history)-1].Escalated = true
					if rerr != nil {
						rec.ReasonerErr = rerr.Error()
						fmt.Printf("      reasoner:   failed — %v\n", rerr)
					} else if rdiag != "" {
						rec.ReasonerFix = rdiag
						fmt.Printf("      reasoner:   %s\n", truncate(rdiag, 160))
						if lastDiag != "" {
							lastDiag += "\n\nREASONER (strong-model diagnosis): " + rdiag
						} else {
							lastDiag = "REASONER (strong-model diagnosis): " + rdiag
						}
					}
				case actionHandoff:
					// The handoff must have something to spend. Deciding to hand
					// off on the final attempt and then stopping means the other
					// model never generates — the same "reachable but unfunded"
					// bug as the escalate ordering. Grant ONE attempt, once.
					if attempt >= attempts && !handoffFunded {
						attempts++
						handoffFunded = true
						fmt.Printf("      budget:     +1 attempt so the handoff can actually draw\n")
					}
					// Same spec, same prose, DIFFERENT model. Nothing about the
					// question changes — only who is asked.
					curModel, curPol, viaHandoff = handoff, handoffPol, true
					rec.HandedOff = true
					fmt.Printf("      handoff:    -> %s (same spec, prover still judges)\n", handoff)
				case actionAbandon:
					attempt = attempts // stop spending on this task
				}
				_ = enc.Encode(rec) // AFTER the decision, so the row carries it
			}
		}
		if passed {
			fmt.Printf("%-20s ==> PROVED\n", stem)
		} else {
			fmt.Printf("%-20s ==> not proved in %d attempts\n", stem, attempts)
		}
	}
	fmt.Println("\nresults:", filepath.Join(runDir, "results-unfettered.jsonl"))
	return nil
}
