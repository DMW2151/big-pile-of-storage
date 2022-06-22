// Output //


output "cmd_cortx_ssh_add" {
  value = "ssh-add ~/.ssh/${module.aws-cortx-instance.cortx_instance_ssh_key}.pem"
}

output "cmd_cortx_mgmt_port_forward" {
  value = "ssh -L 127.0.0.1:28100:mgmt.cortx.internal:28100 ubuntu@${module.aws-cortx-instance.cortx_jump_ip}"
}

output "cmd_cortx_data_port_forward" {
  value = "ssh -L 127.0.0.1:28001:data.cortx.internal:80 ubuntu@${module.aws-cortx-instance.cortx_jump_ip}"
}

output "cmd_cortx_ssh_to_jump" {
  value = "ssh -A -oStrictHostKeyChecking=no ubuntu@${module.aws-cortx-instance.cortx_jump_ip}"
}



// Output: CORTX Server Public IP
output "ref_cortx_jump_ip" {
  value = module.aws-cortx-instance.cortx_jump_ip
}

//
// Output: CORTX Server Internal IP
output "ref_cortx_server_mgmt_ip" {
  value = module.aws-cortx-instance.cortx_server_ip
}

//
// Output: CORTX Storage Server IP
output "ref_cortx_server_data_ip" {
  description = ""
  value       = module.aws-cortx-instance.cortx_server_data_ip
}