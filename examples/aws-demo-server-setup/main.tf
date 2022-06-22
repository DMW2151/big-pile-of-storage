//
// Module to launch a CORTX Server 
//
// Terraform Implementation of Seagate CORTX Instructions (EC2 QuickStart)
// See: https://github.com/Seagate/cortx/blob/main/doc/integrations/AWS_EC2/README.md
//

terraform {

  // [Optional] - TF backend In AWS S3 at a fixed location
  // NOTE FOR USER: UPDATE THIS! OR Replace with your own TF bucket!!
  backend "s3" {
    bucket  = "dmw2151-state"
    key     = "production/cortx-aws-stack.tfstate"
    region  = "us-east-1"
    profile = "default"
  }

  // Provider Versions
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0.0"
    }
  }

  // Terraform Version
  required_version = ">= 1.0.3"

}

// Provider Config //
provider "aws" {
  region = "us-east-1"
}


// Module: aws-cortx-instance - Launches a Demo CORTX Instance
module "aws-cortx-instance" {

  // General
  source = "./../modules/aws-cortx-server"


  // SSH Credentials for Provisioners
  cortx_instance_ssh_key     = "public-jump-1" // NOTE FOR USER: UPDATE THIS!
  cortx_provisioner_user     = "..."           // "root" - In a real configuration DO NOT do this, USE TFVARS
  cortx_provisioner_password = "..."           // "opensource!" - In a real configuration DO NOT do this, USE TFVARS

  // Networking Configuration
  availability_zone = "us-east-1a"

  // Security Group Configuration 
  deployer_ip_blocks = [
    "${chomp(data.http.deployerip.body)}/32",
  ]
}

