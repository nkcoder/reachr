# reachr

AWS network reachability & topology tool — a read-only Go CLI that scans AWS and
renders an interactive whole-VPC topology, plus a reachability engine that explains
**what can actually reach what, and why a path is broken.**

> **Status:** early development. See [DESIGN.md](DESIGN.md) for the v1 design and
> [issue #1](https://github.com/nkcoder/reachr/issues/1) for the slice roadmap.

## Commands (planned)

| Command | Description |
| --- | --- |
| `reachr scan --region <r> [--profile <p>] [--filter tag:project=X]` | Read-only scan → immutable `topology.json` snapshot (the only phase that touches AWS). |
| `reachr render <topology.json> [--format html\|dot]` | Interactive diagram from a snapshot (pure function, no AWS). |
| `reachr explain --from <id> --to <id> --port <n> <topology.json>` | Path + first blocking hop + honest verdict. |
| `reachr diff <old.json> <new.json>` | Snapshot diff over time. |

## Development

```sh
make build   # build the reachr binary
make test    # run tests
make lint    # golangci-lint
```
