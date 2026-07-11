data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }
}

# App tier: two instances (one per AZ) standing in for ECS tasks. They don't need
# to serve traffic — reachr scans structure (ENIs/SGs/target associations), and
# v1 does not evaluate target health.
resource "aws_instance" "app" {
  count                       = 2
  ami                         = data.aws_ami.al2023.id
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.private[count.index].id
  vpc_security_group_ids      = [aws_security_group.app.id]
  associate_public_ip_address = false
  tags                        = { Name = "reachr-testenv-app-${count.index}" }
}

resource "aws_lb" "app" {
  name               = "reachr-testenv-alb"
  internal           = true
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.private[*].id
  tags               = { Name = "reachr-testenv-alb" }
}

resource "aws_lb_target_group" "app" {
  name        = "reachr-testenv-app"
  port        = var.app_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.this.id
  target_type = "instance"
  tags        = { Name = "reachr-testenv-app" }
}

resource "aws_lb_target_group_attachment" "app" {
  count            = length(aws_instance.app)
  target_group_arn = aws_lb_target_group.app.arn
  target_id        = aws_instance.app[count.index].id
  port             = var.app_port
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.app.arn
  port              = 80
  protocol          = "HTTP"
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# DB tier: RDS Postgres (db.t3.micro is RDS free-tier eligible).
resource "random_password" "db" {
  length  = 20
  special = false
}

resource "aws_db_subnet_group" "this" {
  name       = "reachr-testenv"
  subnet_ids = aws_subnet.private[*].id
  tags       = { Name = "reachr-testenv" }
}

resource "aws_db_instance" "this" {
  identifier             = "reachr-testenv"
  engine                 = "postgres"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  db_name                = "reachr"
  username               = "reachr"
  password               = random_password.db.result
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.db.id]
  publicly_accessible    = false
  multi_az               = false
  skip_final_snapshot    = true
  tags                   = { Name = "reachr-testenv-db" }
}
