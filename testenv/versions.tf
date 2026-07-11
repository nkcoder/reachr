terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5"
    }
  }
}

provider "aws" {
  region = var.region

  # Everything is tagged so reachr's scope selector has real data and teardown
  # is trivial (filter/destroy by project=reachr-testenv).
  default_tags {
    tags = {
      project   = var.project_tag
      managedBy = "terraform"
    }
  }
}
