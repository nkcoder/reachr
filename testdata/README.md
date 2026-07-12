# testdata

## testenv.golden.json

A sanitized `topology.json` snapshot of the `testenv/` topology (schemaVersion 1).
This is the golden fixture the reachability engine and renderer are tested against —
no live AWS needed. It exercises the interesting cases:

| | |
|---|---|
| ENIs | 5 — instance (×2), ALB (×2, `amazon-elb`), RDS (`amazon-rds`) |
| Security groups | 4 — the `alb → app → db` SG-references-SG chain + default self-ref |
| Route tables | 2 — main + custom, with a gateway-endpoint prefix-list route |
| VPC endpoints | 1 — S3 gateway endpoint |

**Sanitized:** the AWS account id is redacted to `000000000000` (in `account`, ARNs,
SG-peer accounts, and instance-owner ids). Resource ids and RFC1918 private IPs are
kept — they aren't sensitive and are needed for realistic reasoning. Service owners
(`amazon-elb`, `amazon-rds`) are preserved.

### Regenerating

Stand up `testenv/` (see `testenv/README.md`) and run a scrubbed scan:

```sh
go run . scan --region ap-southeast-2 --scrub \
  --vpc "$(terraform -chdir=testenv output -raw vpc_id)" \
  -o testdata/testenv.golden.json
```

`--scrub` applies the same redaction (`internal/topology.Sanitize`). `scannedAt` will
differ per scan; pin it if a stable diff matters.
