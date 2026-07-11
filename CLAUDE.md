# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

**Pre-implementation.** The repo currently contains only `DESIGN.md` (the v1 design spec) and its
git history. There is no Go module, source code, or build tooling yet. `DESIGN.md` is the source of
truth for scope and architecture — read it before writing code, and keep it in sync when decisions
change. Do not invent build/lint/test commands until the corresponding tooling actually exists.

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
