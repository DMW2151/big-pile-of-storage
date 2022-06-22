//
// Module to Interact with a CORTX Server 
// 

// Terraform Main Config //
terraform {

  // [Optional] - TF backend In AWS S3 at a fixed location
  backend "s3" {
    bucket = "dmw2151-state" // Replace with your TF state bucket (or remove)
    key    = "production/cortx-demo-stack.tfstate"
    region = "us-east-1"
  }

  // Provider Versions
  required_providers {
    cortx = {
      version = "0.0.1"
      source  = "dmw2151.com/terraform/cortx"
    }
  }

  // Terraform Version
  required_version = ">= 1.0.3"

}

// CORTX Provider Config //
provider "cortx" {

  // Sensitive! these variables are hardcoded here, but in a production setup 
  // with a real HOST, ACCESS_KEY, and SECRET_KEY, please apply vars w. an uncommitted
  // *.tfvars file

  // CORTX CloudShare Instance Configuration 
  cortx_endpoint_host = "localhost" // Variable on Cloudshare Instance Start
  cortx_endpoint_port = "28001"     // Constant - Always 31949 in CORTX on CloudShare, In demo, this port forwwards thru 28001

  // CORTX Cloudshare Default Credentials
  cortx_access_key        = "..."
  cortx_secret_access_key = "..."
}

// Resource: TerraformProvisionerTestBucket - Create a new bucket as a Resource
resource "cortx_bucket" "test_provisioner_resource" {
  bucket        = "terraform-provisioner-test-bucket"
  force_destroy = true
}