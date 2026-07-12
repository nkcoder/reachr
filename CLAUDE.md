# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

`DESIGN.md` is the source of truth for scope/architecture — read it before writing code, keep it in
sync when decisions change. Development is tracked on GitHub (repo `nkcoder/reachr`) as **slice-based
milestones**; roadmap is issue **#1**. Progress by slice:

- **Slice 0 — skeleton** ✅ Go module, Cobra CLI, CI, Task runner.
- **Slice 1 — scan + fixture** ✅ `scan` produces a normalized `topology.json`; golden fixture +
  tests committed. Testenv Terraform was applied, scanned, and **destroyed** (rebuild from `testenv/`
  if live AWS is needed again).
- **Slice 2 — structural `dot` render** ⬅️ **NEXT.** `render topology.json --format dot` → Graphviz
  (nodes + subnet clustering + console deep-links), proof = a PNG of the testenv in the README. First
  open design fork: **what is a node** — raw ENIs vs already-aggregated logical resources (ENI→resource
  classification was deliberately deferred to the render layer, so it first bites here).
- Slices 3–9 (SG engine → route engine → explain → labeling → HTML → filter → NACL/polish): see #1.

### Commands

- `task build` / `task test` / `task lint` (Task runner; `brew install go-task`). CI runs the same.
- `go run . scan --region <r> [--profile <p>] [--vpc <id>] [--scrub] [--raw]` — the only AWS-touching
  path. `--raw` dumps raw EC2 JSON; `--scrub` redacts the account id (shareable snapshots / fixtures).
- Tests + render/explain run offline against `testdata/testenv.golden.json` (sanitized real snapshot).

### Key code

- `internal/topology/` — the `topology.json` schema (`types.go`), design rationale (`doc.go`),
  `Sanitize` (`sanitize.go`). `SchemaVersion` is a pre-v1 breakable contract.
- `internal/awsscan/` — AWS SDK v2 client + paginated collectors (`rawdump.go`) and the pure
  `RawBackbone → topology.Snapshot` mapper (`mapper.go`).
- `internal/cli/` — Cobra command tree.
- `testenv/` — free-tier Terraform (private ALB→app→RDS + S3 gw endpoint) that seeds the fixture.

### Schema decisions (from the #7 grill, encoded in `internal/topology/doc.go`)

- **ENI ownership: raw signals stored, classification deferred.** ENI→logical-resource is a heuristic
  (parse `Description`/`RequesterId`/`InstanceOwnerId`), so it belongs in the render/aggregation layer
  (Slice 2/6), not the snapshot.
- **Routes & SG rules ARE normalized in the snapshot** (deterministic unions, not heuristics): routes
  → `{destination, target, state}` (`GatewayId` sub-typed local/igw/vpce; unmapped → `TargetUnknown`);
  SG permissions → flattened atomic rules (`{direction, protocol, ports, peer}`).
- Collections are arrays (engine indexes at load); tags are maps; IPv6 stored but not reasoned in v1.

### Workflow conventions

- One slice at a time; small PR per issue on a branch `slice-N/<topic>`; every PR must be CI-green.
- Use **`Closes #N`** in the PR body only when it fully finishes an issue (else `Part of #N`).
- Detail issues just-in-time per slice; don't over-plan far slices. Keep `raw.json`/`vpce.raw.json`
  out of git (gitignored, unsanitized).
- GitHub Project #3 board: the token lacks `project` scope, so issues aren't auto-added via CLI
  (user wires the board via the Auto-add workflow).

## What this is

`reachr` — a read-only Go CLI that scans AWS and produces (a) an interactive whole-VPC topology
diagram and (b) a reachability engine answering "what can actually reach what, and why is a path
broken." Positioned against Cloudmapper/Cartography/inframap/VPC Reachability Analyzer; the gap it
fills is a lightweight live-scan, whole-topology, honest-partial-fidelity CLI.

## Architecture — the load-bearing decisions

These are the invariants that span multiple future files. Preserve them; if you deviate, update
`DESIGN.md` and say why.

- **Two-phase pipeline.** `scan` (needs AWS creds) writes an **immutable `topology.json` snapshot**;
  `render` / `explain` / `diff` are **pure functions of that snapshot with no AWS calls**. This
  split is the basis of the whole testing strategy — do not let credential-dependent logic leak into
  the render/explain phase.

- **Model at ENI level, render at logical-resource level.** The internal graph backbone is ENIs:
  `ENI → (SG rules + route tables) → ENI`. Rendering aggregates ENIs up to logical resources (one
  node per ECS service / RDS / ALB / NAT; an ASG of N identical tasks → one node labeled "×N"). SGs
  collapse into **edge properties, not nodes**.

- **Honest partial fidelity.** Every reachability verdict must declare what it did and did **not**
  evaluate (e.g. `✅ SG allows  ✅ route exists  ⚠️ NACL/DNS/target-health not evaluated`). A
  confidently-wrong answer is worse than none. v1 reasons about SG (rung 1) + route tables (rung 2) +
  light NACL (rung 3); peering/TGW/DNS/app-layer are deferred and must be **loudly labeled**, never
  silently assumed.

- **Demote, don't delete.** Dead/orphaned/noisy resources are collapsed into a labeled "detached"
  cluster plus a side list — never silently hidden. Filtering is layered: liveness filter (default
  on) → tag scope selector (`--filter tag:project=X`, the main knob) → connectivity dimming →
  interactive toggles (orphans/untagged off by default).

- **Collect broadly, reason narrowly.** The snapshot captures more than v1's engine reasons about
  (e.g. Route 53 private zones are collected but not resolved). Adding data to the snapshot is not the
  same as claiming to model it.

- **Scan boundary:** single region + single account per scan, explicitly labeled. Cross-account /
  global targets (CloudFront, Global Accelerator, public R53, PrivateLink far side) are drawn as
  labeled **boundary stubs** with edges preserved but internals not claimed. Global services
  (CloudFront / global R53 / global WAF) are scanned from **us-east-1** regardless of `--region`.

## Command surface (planned)

- `scan --region --profile [--filter tag:project=X]` → writes `topology.json`
- `render topology.json` → self-contained interactive HTML (cytoscape.js, no frontend build / no
  server); `--format dot` is the Graphviz escape hatch. Every node deep-links to the AWS console via
  its resource ID.
- `explain --from <resource> --to <resource> --port <n>` (stretch) → path + first blocking hop +
  owning repo (from tags).
- `diff old.json new.json` (later).

## Auth

AWS SDK for Go v2 **default credential chain** only (profiles/SSO/env/assume-role) — no custom auth
layer. A read-only role is the user's responsibility.

## Testing strategy

The render/explain engine is a pure function of the snapshot, so test it against **recorded/crafted
fixture snapshots** — no live AWS. The canonical test shape: craft a `topology.json` with a
deliberately misconfigured SG, assert the path-explainer names the correct blocking hop. This is how
verdicts earn trust; prioritize it when the engine lands.
