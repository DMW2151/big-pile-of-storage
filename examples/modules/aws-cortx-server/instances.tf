//
// CORTX Server Instance
//

// AMI for Seagate CORTX Server
// Data: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ami
data "aws_ami" "cortx" {

  // General Search Params
  owners      = ["self"]
  most_recent = true

  // Select most recent CORTX version *you've* imported into AWS (1.0.4)
  filter {
    name   = "name"
    values = ["seagate_cortx_ova_1_0_4"] // Depends on AMI Import Name...
  }

  // Virt Type - Always HVM
  // See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/virtualization_types.html
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

}

// Data: ...
data "aws_ami" "ubuntu" {

  // General Search Params
  owners      = ["099720109477"] // Canonical
  most_recent = true

  // Select most recent CORTX version *you've* imported into AWS (1.0.4)
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  // Virt Type - Always HVM
  // See: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/virtualization_types.html
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

}



// Seagate CORTX Server Instance
//
// NOTE: Because of the Way CORTX assigns ENS on Boot; there is an 
// explicit dependency on the attached ENI
//
// - aws_network_interface.cortx_ens_32
// - aws_network_interface.cortx_ens_33
//
// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance 
resource "aws_instance" "cortx" {

  // General
  ami               = data.aws_ami.cortx.id
  instance_type     = "c5.large" // Hardcodedd the Instance Type for Compliance w. AWS setup
  availability_zone = var.availability_zone
  subnet_id         = aws_subnet.public.id

  // Security Configuration for Public ENI
  vpc_security_group_ids = [
    aws_security_group.vpc_all_traffic.id,
    aws_security_group.public_internet_egress.id,
    aws_security_group.allow_deployer_all_traffic.id,
  ]

  // SSH Configuration
  key_name = var.cortx_instance_ssh_key

  // Block Storage – EBS optimized is default on *most* suitable instances
  ebs_optimized = true

  // Monitoring
  monitoring = true

  // Routing
  associate_public_ip_address = true // No need, except for provisioning - again, not safe!

  // Tags
  tags = {
    Environment = "Production"
    Owner       = "Terraform"
    Name        = "CORTX Server"
  }

}


// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance 
resource "aws_instance" "jump" {

  // General
  ami               = data.aws_ami.ubuntu.id
  instance_type     = "t3.micro"
  availability_zone = var.availability_zone
  subnet_id         = aws_subnet.public.id

  // Security Configuration for Public ENI
  vpc_security_group_ids = [
    aws_security_group.vpc_all_traffic.id,
    aws_security_group.allow_deployer_all_traffic.id,
    aws_security_group.public_internet_egress.id,
  ]

  // SSH Configuration
  key_name = var.cortx_instance_ssh_key

  // User Data
  user_data = file("${path.module}/cloud-init/ec2-provision-client.sh")

  // Tags
  tags = {
    Environment = "Production"
    Owner       = "Terraform"
    Name        = "CORTX Jump Server"
  }

}
