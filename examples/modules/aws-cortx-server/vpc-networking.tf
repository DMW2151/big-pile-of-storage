//
// AWS VPCs and Subnets - Initialize a VPC w. very light networking
// just enough to have CORTX up and running
//
//  - VPC - A small VPC in AWS provider's region
//  - Subnets - A single subnet - US-EAST-1A by default - just need a place for the instance
//  - IGW - Allow internet access to the instance
//  - Network Interfaces - As described by CORTX install guide, need 2 additional ENIs
//
//

// VPC //

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc
resource "aws_vpc" "cortx" {

  // General
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  // Tags
  tags = {
    Name        = "CORTX VPC"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// VPC Routing //

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/internet_gateway
resource "aws_internet_gateway" "cortx_igw" {

  // VPC
  vpc_id = aws_vpc.cortx.id

  // Tags
  tags = {
    Name        = "CORTX VPC - IGW"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table
resource "aws_route_table" "main" {

  // General
  vpc_id = aws_vpc.cortx.id

  // Routes All Addresses, both IPV4 and IPV6 from Internet to IGW
  //
  // Terraform Docs:
  //
  // Note that the default route, mapping the VPC's CIDR block to "local", is
  // created implicitly and cannot be specified.

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.cortx_igw.id
  }

  route {
    ipv6_cidr_block = "::/0"
    gateway_id      = aws_internet_gateway.cortx_igw.id
  }


  // Tags
  tags = {
    Name        = "CORTX VPC - Main Route Table"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/main_route_table_association
resource "aws_main_route_table_association" "asc_main_vpc" {
  vpc_id         = aws_vpc.cortx.id
  route_table_id = aws_route_table.main.id
}


// Secondary - Explicit Route Table Associations //

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table_association
resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.main.id
}


// Subnets //

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet
resource "aws_subnet" "public" {

  // General
  vpc_id                  = aws_vpc.cortx.id
  cidr_block              = cidrsubnet(aws_vpc.cortx.cidr_block, 4, 0)
  availability_zone       = var.availability_zone
  map_public_ip_on_launch = true

  // Tags
  tags = {
    Name        = "CORTX Public Subnet - 1"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// ENI - EIP //

// NOTE - These are **additional** interfaces on top of the instance's default 1 ENI. Add these for
// exposing other mgmt endpoints (?)

// First Network Interface - Refer to CORTX Install Guide for Discussion
// https://github.com/Seagate/cortx/blob/main/doc/integrations/AWS_EC2/README.md
//
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/network_interface
resource "aws_network_interface" "cortx_ens_33" {

  // General
  subnet_id   = aws_subnet.public.id
  description = "cortx_network_interface_33"

  // Security Group Configuration
  security_groups = [
    aws_security_group.vpc_all_traffic.id,
    aws_security_group.allow_deployer_all_traffic.id,
  ]

  // Attach to CORTX Server 
  attachment {
    instance     = aws_instance.cortx.id
    device_index = 1
  }

  // Tags
  tags = {
    Name        = "cortx_network_interface_33"
    Description = "cortx_network_interface_33"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// Second Network Interface - Refer to CORTX Install Guide for Discussion
// https://github.com/Seagate/cortx/blob/main/doc/integrations/AWS_EC2/README.md
//
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/network_interface
resource "aws_network_interface" "cortx_ens_32" {

  // General
  subnet_id   = aws_subnet.public.id
  description = "cortx_network_interface_32"

  // Security Group Configuration
  security_groups = [
    aws_security_group.vpc_all_traffic.id,
    aws_security_group.allow_deployer_all_traffic.id,
  ]

  // Attach to CORTX Server 
  attachment {
    instance     = aws_instance.cortx.id
    device_index = 2
  }

  // Tags
  tags = {
    Name        = "cortx_network_interface_32"
    Description = "cortx_network_interface_32"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


