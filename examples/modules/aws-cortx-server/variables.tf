//
// Variables - See Demo README.MD for detail
//

// Variable: cortx_provisioner_user
variable "cortx_provisioner_user" {
  type = string
}

// Variable: cortx_provisioner_password
variable "cortx_provisioner_password" {
  type = string
}

// Variable: cortx_instance_ssh_key
variable "cortx_instance_ssh_key" {
  type    = string
  default = "public-jump-1"
}

// Variable: Availability Zone
variable "availability_zone" {
  type    = string
  default = "us-east-1a"
}

// Variable: deployer_ip_blocks - Yout network'S IP or CIDR range
variable "deployer_ip_blocks" {
  type      = list(string)
  sensitive = true
}