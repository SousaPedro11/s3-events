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
    bucket       = "your-unique-tf-state-bucket-name" # Change to your unique S3 bucket for Terraform state
    key          = "s3-events.tfstate"
    region       = "us-east-1" # Change to the region of your S3 bucket
    use_lockfile = true
  }
}

provider "aws" {
  region = "us-east-1" # Change to your desired region
}

