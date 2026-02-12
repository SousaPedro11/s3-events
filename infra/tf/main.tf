terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4"
    }
  }

  backend "s3" {
    bucket       = ""
    key          = ""
    region       = ""
    profile      = ""
    use_lockfile = true
  }
}

provider "aws" {
  region = "us-east-1" # Change to your desired region
}

