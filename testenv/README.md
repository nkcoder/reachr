# testenv — free-tier private topology

A small, **fully private** topology (Terraform) that gives `reachr` real ENIs, a
three-tier security-group chain, and a VPC endpoint to design the `topology.json`
schema against and to seed the golden fixture. Not part of the shipped tool —
it's a development/test environment.

```
Custom VPC 10.20.0.0/16 (no IGW / no NAT)
 └─ 2 private subnets (2 AZs)
     ├─ internal ALB [alb-sg] :80 ─▶ target group :8080
     │        └─ 2× EC2 t2.micro [app-sg]      (stands in for ECS)
     ├─ RDS db.t3.micro Postgres [db-sg]
     └─ S3 gateway endpoint (free) → prefix-list route
SG chain:  alb-sg ─:8080▶ app-sg ─:5432▶ db-sg   (SG-references-SG)
```

## Cost

Designed for the **12-month AWS Free Tier**: EC2 `t2.micro`, RDS `db.t3.micro`,
and one ALB are each within the 750 hrs/month allowance; the S3 gateway endpoint
is always free. There is **no NAT gateway or interface endpoint** (the usual cost
drivers). Still — **run `terraform destroy` once you've captured the fixture.**
Two app instances share the 750-hr EC2 allowance, so don't leave it running.

## Usage

```sh
cd testenv
terraform init
terraform apply          # ~10 min (RDS is the slow part)

# Scan it (from the repo root, using the scan --raw command from Slice 1):
cd ..
go run . scan --raw --region ap-southeast-2 --vpc "$(terraform -chdir=testenv output -raw vpc_id)" > raw.json

# When done:
terraform -chdir=testenv destroy
```

`terraform output scan_command` prints the exact scan command with the VPC id
filled in.

## Notes

- Apps don't serve traffic; ALB targets will show unhealthy. That's fine — reachr
  scans **structure**, and v1 doesn't evaluate target health.
- The RDS master password is generated (`random_password`) and never needed; the
  DB is not reachable and exists only for its network shape.
