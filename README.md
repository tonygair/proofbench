# proofbench — repeat runs for AI-generated verified code

A small harness that asks a **local** model to write SPARK Ada bodies against 58
fixed specifications, and lets a **prover** decide whether each one holds.

I am asking for repeat runs on other people's hardware. The reason is in
`WHY_REPEATS.md`, and it is not a formality — one draw is a sample, and I have the
numbers to show it.

---

## Before anything else: what this does, and what it does not

You are being asked to run a stranger's code. Fair enough. Everything below is
checkable by reading, and the whole harness is about 1,500 lines of Go.

**It does:**
- start a local process that talks to **your own ollama** on `127.0.0.1:11434`
- run `gnatprove` on your machine to judge each candidate
- write results to a `.jsonl` file in the working directory

**It does not:**
- make any outbound network request. The only HTTP call in the source is to the
  `-rail` address, which defaults to `127.0.0.1`. Grep for `http` and check —
  there are three occurrences and they are all that one call.
- read anything outside its own directory except the two tools it invokes
  (`ollama` via HTTP, `gnatprove`/`gnatchop` via `PATH`)
- write anywhere except its working directory
- send me anything. There is no telemetry, no key, no account, nothing to sign up
  for. If you want to share a result you send me the file yourself, or you don't.

**Audit it in ten minutes:**

    grep -rn "http\|exec.Command\|os.Create\|os.WriteFile" harness/*.go

That is the complete list of everything it talks to and everything it writes.

---

## What you need

- **ollama** with a code model pulled (`ollama pull qwen3-coder:30b` — about 18 GB)
- **gnatprove + gnatchop**, easiest via [Alire](https://alire.ada.dev):
  `alr toolchain --select` then `alr install gnatprove`
- **Go** to build the harness
- **Python 3** — only to run the judge script, which came from the upstream
  benchmark (see Credits) and is 400 readable lines

## Run it

    cd harness && go build -o proofbench .
    cd ..
    ALL=$(ls tasks/*.ada | sed 's|tasks/||;s|\.ada$||' | paste -sd, -)
    harness/proofbench -track . -run harness -only "$ALL" \
        -model qwen3-coder:30b -attempts 1 -timeout 3m > run1.log

Then again, twice more, into `run2.log` and `run3.log`. **Three runs is the point** —
one is not useful to me and would not be useful to you either.

    grep -c '==> PROVED' run1.log

## What I would like back

The dull stuff, and nothing else:

- the three scores
- model name and quantisation
- GPU / CPU and RAM
- the `.log` files if you are willing

No system information beyond that, and I have no way to collect any of it myself.

---

## The one thing that will surprise you

Your three scores will not match each other. Mine did not: **30, 31, 32 and 34**
across four runs of an identical configuration on 58 tasks.

That is why I am asking. See `WHY_REPEATS.md`.

**And it is worse than that, which is the honest part.** A later series of six
runs on the same tasks and the same model gave **26, 26, 26, 26, 28 and 29** —
a tight spread, but well below the first series. Between the two, the harness
changed: portability fixes, a manual-lookup fallback, and the machine was busy
with other work. I have not isolated which of those moved it, so I am not going
to tell you the two series are comparable.

⇒ The shipped `harness/results-unfettered.jsonl` is from the SECOND series, so
its numbers will not match the first ones quoted above. That is not an error in
the file. It is the thing this whole repository is about: **a score belongs to
the exact configuration and conditions that produced it, and I cannot show you
the first series' conditions any more.**

## Known-impossible tasks — do not count these against your model

Four of the 58 have specifications that **cannot be satisfied by any
implementation**, in any language. They are kept exactly as published so results
stay comparable, but they will never pass:

`NpIntersect` · `NpArange` · `NpDiagonal` · `NpGcd`

A fifth, `NpCumProd`, is satisfiable but needs an identity the prover does not
have, so a correct implementation still fails.

⇒ **The achievable ceiling is 54, not 58.**

## Credits and licence

The task specifications are translated from the `numpy_simple` set of
**[vericoding](https://github.com/Beneficial-AI-Foundation/vericoding)** by the
Beneficial AI Foundation (MIT). `judge/spark_verify.py` is theirs. The translation
to SPARK and the harness are mine, under the same licence. Their benchmark being
public and well documented is the only reason any of this exists.
