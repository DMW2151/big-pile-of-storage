// Security Groups for AWS Demo 

// Allows Full Access from the IP of the System Deployer/Admin
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group
resource "aws_security_group" "allow_deployer_all_traffic" {

  // General
  name                   = "cortx_allow_all_deployer"
  description            = "Allows All Connections from Deployer"
  vpc_id                 = aws_vpc.cortx.id
  revoke_rules_on_delete = true

  // Egress / Ingress Rules - Allow All Traffic From
  // Deployer Instance
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = var.deployer_ip_blocks
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = var.deployer_ip_blocks
  }

  // Tags
  tags = {
    Name        = "CORTX Allow Deployer All Traffic"
    Owner       = "Terraform"
    Environment = "Production"
  }
}


// An [overly] permissive group that allows all communication within the VPC,
// a safer configuration may involve limiting this to
//
//  - Allows HTTP, HTTPS communications from within VPC
//  - Allow CORTX Management Comms - TCP on 28100 From Within VPC
//  - Allow SSH within VPC
//  - Allow ICMP Ping
//
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group
resource "aws_security_group" "vpc_all_traffic" {

  // General
  name                   = "cortx_vpc_all_traffic"
  description            = "Allows ingress/egress on all ports from within the VPC"
  vpc_id                 = aws_vpc.cortx.id
  revoke_rules_on_delete = true

  // Ingress/Egress Rules - Allow all connections within the VPC
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.cortx.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.cortx.cidr_block]
  }

  // Tags
  tags = {
    Name        = "CORTX Allow Intra-VPC All Traffic"
    Owner       = "Terraform"
    Environment = "Production"
  }

}


// Allow Egress to the public internet - If configured properly, through
// `
//
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group
resource "aws_security_group" "public_internet_egress" {

  // General
  name                   = "cortx_public_internet"
  description            = "Allows Egress to Public Internet"
  vpc_id                 = aws_vpc.cortx.id
  revoke_rules_on_delete = true

  // Egress
  egress {
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  // Tags
  tags = {
    Name        = "CORTX Allow Egress to Internet"
    Owner       = "Terraform"
    Environment = "Production"
  }

}
