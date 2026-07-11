# Three-tier SG chain: alb -> app -> db, using SG-references-SG rules (the
# rung-1 reachability engine's key case).

resource "aws_security_group" "alb" {
  name        = "reachr-testenv-alb"
  description = "ALB tier"
  vpc_id      = aws_vpc.this.id
  tags        = { Name = "reachr-testenv-alb" }
}

resource "aws_security_group" "app" {
  name        = "reachr-testenv-app"
  description = "App tier (stands in for ECS)"
  vpc_id      = aws_vpc.this.id
  tags        = { Name = "reachr-testenv-app" }
}

resource "aws_security_group" "db" {
  name        = "reachr-testenv-db"
  description = "RDS tier"
  vpc_id      = aws_vpc.this.id
  tags        = { Name = "reachr-testenv-db" }
}

# ALB: HTTP in from within the VPC (no external client in a fully private env).
resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  description       = "HTTP from within VPC"
  cidr_ipv4         = var.vpc_cidr
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "alb_to_app" {
  security_group_id            = aws_security_group.alb.id
  description                  = "To app tier"
  referenced_security_group_id = aws_security_group.app.id
  from_port                    = var.app_port
  to_port                      = var.app_port
  ip_protocol                  = "tcp"
}

# App: in from ALB on app_port (SG-references-SG).
resource "aws_vpc_security_group_ingress_rule" "app_from_alb" {
  security_group_id            = aws_security_group.app.id
  description                  = "From ALB"
  referenced_security_group_id = aws_security_group.alb.id
  from_port                    = var.app_port
  to_port                      = var.app_port
  ip_protocol                  = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "app_to_db" {
  security_group_id            = aws_security_group.app.id
  description                  = "To DB"
  referenced_security_group_id = aws_security_group.db.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
}

# DB: Postgres in from app (SG-references-SG).
resource "aws_vpc_security_group_ingress_rule" "db_from_app" {
  security_group_id            = aws_security_group.db.id
  description                  = "Postgres from app"
  referenced_security_group_id = aws_security_group.app.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
}
