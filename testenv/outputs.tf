output "vpc_id" {
  description = "The test VPC id"
  value       = aws_vpc.this.id
}

output "scan_command" {
  description = "Ready-to-run exploratory scan for this VPC (run from the repo root)"
  value       = "go run . scan --raw --region ${var.region} --vpc ${aws_vpc.this.id} > raw.json"
}

output "alb_dns_name" {
  description = "Internal ALB DNS name"
  value       = aws_lb.app.dns_name
}
