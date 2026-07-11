# AWS Network Reachability & Topology Tool — Design v1

A Go CLI that scans AWS (read-only) and produces an interactive whole-topology
diagram plus a reachability engine that answers **"what can actually reach what,
and why is this path broken."**

> Origin: started as a "Well-Architected resource scanner" idea, but that space is
> saturated (Prowler, Powerpipe WAF mod, Trusted Advisor). Pivoted to network
> reachability — a real, under-served pain, especially when a project's topology is
> spread across 10+ Terraform repos and the only complete picture lives in live AWS.

## Goals

- **Primary:** learning / portfolio / OSS project with genuine practical value at work.
- Understand a whole VPC's connectivity at a glance (diagram-first).
- Debug "why can't X reach Y" with an honest, evidence-backed answer.
- Only requirement: a **read-only role** on the account.

## Prior art & the gap we fill

- **Cloudmapper** (Duo) — did this, now archived/unmaintained.
- **AWS Perspective / Workload Discovery** — official but heavy (deploys a stack).
- **Cartography** (Lyft) — graph DB, not a "network flow" diagram.
- **inframap** — from Terraform state, not live; misses drift + merged multi-repo reality.
- **VPC Reachability Analyzer** — pairwise only, pay-per-analysis, no whole-VPC view.
- **The gap:** lightweight CLI, live read-only scan, whole-topology connectivity view,
  resource IDs deep-linking to the console, honest partial reachability.

## Core principles

1. **Reachability, not just inventory.** Model what can _actually_ talk to what.
2. **Honest partial fidelity.** A confidently-wrong debugging tool is worse than none.
   Every verdict declares what it did and did **not** evaluate.
3. **Demote, don't delete.** Noise is collapsed/dimmed and reported, never silently hidden.
4. **Collect broadly, reason narrowly.** Snapshot captures more than v1's engine reasons about.
5. **Model at ENI level, render at logical-resource level.**

## Architecture — two-phase pipeline

`scan` (needs AWS creds) → **immutable `topology.json` snapshot** → `render` / `explain`
/ `diff` (pure functions of the snapshot, no creds).

Buys us: testability (replay fixture snapshots — no live AWS), fast offline iteration,
diff-over-time for free, and shareable snapshots for teammates without AWS access.

### The graph model

- **Backbone = ENIs.** Everything network-attached in AWS hangs off an ENI; SGs attach
  to ENIs. Internal graph is `ENI → (SG rules + route tables) → ENI`.
- **Render** aggregates ENIs up to _logical resources_ (one node per ECS service / RDS /
  ALB / NAT). An ASG of 10 identical tasks → one node labeled "×10". SGs collapse into
  edge properties, not nodes.
- **Hierarchical grouping:** VPC → AZ/subnet → resource, collapsible.

### Reachability fidelity ladder (v1 = rungs 1–2 + light 3)

1. **Security groups** — stateful, ingress/egress, SG-references-SG, prefix lists. ✅ v1
2. **Route tables** — per-subnet, longest-prefix, blackhole, IGW/NAT/egress-only. ✅ v1
3. **NACLs** — stateless, ordered, ephemeral return ports. ⚠️ light in v1
4. VPC endpoints / endpoint policies — _partly in v1 (see scope)_
5. Peering / Transit Gateway — ❌ deferred (route propagation, non-transitivity)
6. DNS (Route 53 private zones, resolver rules) — ❌ deferred (collected, not resolved)
7. App layer (ALB/NLB listeners, target health, cross-account PrivateLink) — ❌ deferred

Every path verdict prints, e.g.: `✅ SG allows  ✅ route exists  ⚠️ NACL/DNS/target-health not evaluated`.

## Filtering — layered relevance model (demote, don't delete)

1. **Liveness filter** (default on): collapse objectively-dead into a "detached" cluster
   _and_ emit a side list — unattached ENIs (`status=available`), SGs referenced by nothing,
   blackhole routes, unassociated EIPs, empty target groups. (Side list = cost/cleanup bonus.)
2. **Scope selector** (opt-in, the main knob): `--filter tag:project=X`. Tagging is
   consistent in the target environments (`project` / `ownedBy` / `environment` / `controlledBy`),
   so this is a clean tag query, not a heuristic. Tag-matched resources = "the core".
3. **Connectivity dimming:** anything not in the core's connected component → side cluster.
4. **Interactive toggles:** "show orphans" / "show untagged" — off by default, one click away.

## Scan boundary (v1)

- **Single region** per scan (`--region`), explicitly labeled. Multi-region merge later.
- **Single account.** Cross-account targets (PrivateLink far side, cross-account peering)
  drawn as labeled **boundary stubs** ("external / acct 1234…"), edges preserved, internals
  not claimed.
- **Global services:** collect broadly, reason narrowly:
  - _Draw as ingress/boundary stubs:_ CloudFront, Global Accelerator, public Route 53 records
    → your ALB/CloudFront front door.
  - _Collect into snapshot but don't resolve:_ Route 53 **private** hosted zones (banks the
    data for the deferred DNS rung).
  - _Ignore:_ IAM et al. (not network-relevant).
  - Note: CloudFront / global R53 / global WAF are served from **us-east-1** — scan
    "regional from `--region` + global always from us-east-1", labeled as global.

## v1 scope (tracer bullet)

**Resource coverage — the core data-plane path:**

- VPC, subnets (+AZ), route tables, IGW/NAT
- **ENIs** (backbone), **security groups** (edges), NACLs (light)
- **ALB/NLB** + target groups → targets
- **ECS** (services/tasks), **RDS**
- **VPC endpoints — interface (PrivateLink) + gateway (S3/DDB)**; PrivateLink near-side
  (far side = boundary stub). Interface endpoints are ~free: they ride the ENI+SG model;
  extra work is just labeling the service name. Central to private-subnet ECS setups.
- Tags on everything (scope selector).

**Deferred to v1.1+ (loudly labeled "not yet modeled"):**

- VPC peering, Transit Gateway (the real effort spike — route propagation/non-transitivity)
- Private API Gateway, Lambda-in-VPC
- Route 53 private-zone DNS resolution (rung 6)
- Multi-region merge, cross-account read

## Rendering

- **Primary output:** self-contained **interactive HTML** (cytoscape.js) — VPC/subnet
  grouping, filtering, click-to-expand, each node **deep-links to the AWS console** via its
  resource ID. Go templates JSON into a bundled HTML shell (single binary, no frontend build,
  no server).
- **Escape hatch:** `--format dot` (Graphviz) for a quick static picture / piping.

## Command surface

- `scan --region --profile [--filter tag:project=X]` → writes `topology.json`
- `render topology.json` → interactive HTML (`--format dot` escape hatch)
- _(stretch)_ `explain --from <resource> --to <resource> --port <n>` → path + **first
  blocking hop** + **owning repo** (from tags) — the multi-repo debugging story.
- _(later)_ `diff old.json new.json`

## Auth

- AWS SDK for Go v2 **default credential chain** (profiles, SSO, env, assume-role — free).
- No custom auth layer. Document the minimum read-only permissions
  (`ec2:Describe*`, `elasticloadbalancing:Describe*`, `rds:Describe*`, `ecs:Describe*/List*`,
  `route53:List*/Get*`, tag read). Read-only role is the user's responsibility.

## Testing strategy

- Engine (`render`/`explain`) is a **pure function of the snapshot** → unit-test against
  **recorded/crafted fixture snapshots**. Craft a snapshot with a deliberately misconfigured
  SG; assert the path-explainer names the correct blocking hop. This is how verdicts earn trust.

## Definition of done — v1

Given a read-only role, `scan` a real VPC → `render` produces an interactive HTML topology
showing ALB → ECS → RDS **+ VPC/PrivateLink endpoints**, with SG/route reachability edges,
tag-scoped to one project, dead/orphaned resources demoted to a labeled cluster + side list,
and every node deep-linking to the console.
