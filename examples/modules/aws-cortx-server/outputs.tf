//
// Module Outputs - Used to Construct Developer Commands at Root Module Level
//

// Output: CORTX STORAGE SERVER IP - MGMT
output "cortx_server_ip" {
  value = aws_instance.cortx.private_ip
}

// Output: CORTX STORAGE SERVER IP - DATA
output "cortx_server_data_ip" {
  value = aws_network_interface.cortx_ens_33.private_ip
}

// Output: CORTX JUMP SERVER IP
output "cortx_jump_ip" {
  value = aws_instance.jump.public_dns
}

// Output: CORTX_INSTANCE_SSH_KEY
output "cortx_instance_ssh_key" {
  value = var.cortx_instance_ssh_key
}