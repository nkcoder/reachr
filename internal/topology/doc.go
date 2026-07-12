// Package topology defines the reachr snapshot schema: the immutable topology.json
// contract that `scan` writes and `render`/`explain`/`diff` read as a pure function.
//
// Design decisions (v1, schemaVersion 1), made against real scanned data:
//
//   - Curated-generous, our own types. We capture the network-relevant fields plus a
//     margin (more than the v1 engine reasons about), not raw AWS SDK structs.
//
//   - ENI ownership is NOT classified here. Every network-attached resource hangs off
//     an ENI, but which logical resource owns an ENI (instance / ALB / RDS / endpoint)
//     is a heuristic over Attachment + Description, so we store the raw signals and defer
//     classification to the aggregation layer (render), where it is a testable pure fn.
//
//   - Routes and SG rules ARE normalized here, because they are deterministic
//     discriminated unions (which AWS field is set determines the meaning), not
//     heuristics. Routes collapse to {destination, target, state}; unmapped targets
//     become TargetUnknown so the engine can flag them honestly rather than drop them.
//     SG permissions flatten to one atomic rule per peer — the engine's evaluation unit.
//
//   - Collections are arrays; the engine builds id->object indices at load. Tags are
//     maps for O(1) scope-selector lookups. IPv6 fields are stored but not reasoned
//     about in v1 (collect broadly, reason narrowly).
//
// schemaVersion is an internal contract, not a public API: pre-v1 it may change freely
// on a version bump.
package topology
