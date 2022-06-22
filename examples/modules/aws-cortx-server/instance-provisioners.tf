// 
// Provisioners
//
// Not the best pattern for Terraform - A series of sequentially executed
// remote-exec steps to provision a test CORTX instance. Include the following:
//
// - Assign Network Interfaces
// - Boostrap CORTX Node and Start all Services (Consul, ES, S3IAMCLI, etc.)
// - Open Hole in Firewalld for testing
// - Create S3 Accounts and Users
// 


// Interface Provisioner - Assign permanent interface to ENS32 / ENS33 as defined in 
// `vpc-networking.tf`, requires a shutdown
//
// Resource: https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource
resource "null_resource" "interfaces" {

  // Connection
  connection {
    type        = "ssh"
    user        = var.cortx_provisioner_user
    password    = var.cortx_provisioner_password
    host        = aws_instance.cortx.public_ip
    private_key = file("~/.ssh/${aws_instance.cortx.key_name}.pem") // Assumes ~/.ssh/**
  }

  // Remote Provisioner - Assign Interfaces
  provisioner "remote-exec" {
    script = "${path.module}/cloud-init/ec2-assign-interfaces.sh"
  }

  // Trigger - Changes to CORTX Instance Force Full Re-Provisioning
  triggers = {
    cortx_instance_id = "${aws_instance.cortx.id}"
  }

}


// Wait - Wait for 60s, there's a shutdown following the reassignment of network interfaces
// See: `null_resource.interfaces` script
//
// Resource: https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource
resource "time_sleep" "wait_60_seconds" {
  create_duration = "60s"
  depends_on      = [null_resource.interfaces]
}


// Cortx Boostrap - Run the CORTX Bootstrap Script. Involves the full CORTX service, spinning up 
// MOTR, Consul, ElasticSearch, etc. ETA: ~5 min.
//
// Resource: https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource
resource "null_resource" "bootstrap" {

  // Connection
  connection {
    type        = "ssh"
    user        = var.cortx_provisioner_user
    password    = var.cortx_provisioner_password
    host        = aws_instance.cortx.public_ip
    private_key = file("~/.ssh/${aws_instance.cortx.key_name}.pem")
  }

  // Remote Provisioner - Spin Up CORTX Instance Services
  provisioner "remote-exec" {
    script = "${path.module}/cloud-init/ec2-bootstrap-cortx-node.sh"
  }

  // Trigger - Changes to CORTX Instance Force Full Re-Provisioning
  triggers = {
    cortx_instance_id = "${aws_instance.cortx.id}"
  }

  // Depends On Wait - Need network interfaces assigned in `null_resource.interfaces`
  // give this 60s to tolerate a reboot
  depends_on = [time_sleep.wait_60_seconds]
}


// CORTX Dummy Users - Creates a Dummy Account and Dummy User for CORTX
// Resource: https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource
resource "null_resource" "init_terraform_user" {

  // Connection
  connection {
    type        = "ssh"
    user        = var.cortx_provisioner_user
    password    = var.cortx_provisioner_password
    host        = aws_instance.cortx.public_ip
    private_key = file("~/.ssh/${aws_instance.cortx.key_name}.pem")
  }

  // Trigger - Changes to CORTX Instance Force Full Re-Provisioning
  triggers = {
    cortx_instance_id = "${aws_instance.cortx.id}"
  }

  // Remote Provisioner - Create CORTX Account and User
  provisioner "remote-exec" {
    script = "${path.module}/cloud-init/ec2-init-cortx-user.sh"
  }

  // Depends On ...
  depends_on = [null_resource.bootstrap]

}

// Remove Firewalld - Allow easier module testing and deployment by stopping Firewalld (started)
// by the boostrap step
//
// WARN: This is a security gap! Don't do in Production! This is why we have security group rules, though!
//
// Resource: https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource
resource "null_resource" "open_networking" {

  // Connection
  connection {
    type        = "ssh"
    user        = var.cortx_provisioner_user
    password    = var.cortx_provisioner_password
    host        = aws_instance.cortx.public_ip
    private_key = file("~/.ssh/${aws_instance.cortx.key_name}.pem")
  }

  // Trigger - Changes to CORTX Instance Force Full Re-Provisioning
  triggers = {
    cortx_instance_id = "${aws_instance.cortx.id}"
    jump_instance_id  = "${aws_instance.jump.id}"
  }

  // Remote Provisioner - Open FirewallD
  // Ref: https://wbhegedus.me/avoiding-asymmetric-routing/
  provisioner "remote-exec" {
    inline = [
      "sudo firewall-cmd --zone=public-data-zone --add-port=443/tcp",
      "sudo firewall-cmd --zone=public-data-zone --add-port=80/tcp",
      "sudo firewall-cmd --zone=public-data-zone --remove-icmp-block=echo-request",
      "sudo ip route add ${aws_network_interface.cortx_ens_33.private_ip} dev ens33 table 1",             // One Addr
      "sudo ip route add ${aws_subnet.public.cidr_block} dev ens33 table 1",                              // Entire Subnet
      "sudo ip route add default via ${aws_network_interface.cortx_ens_33.private_ip} dev ens33 table 1", // Default
      "sudo ip rule add from ${aws_network_interface.cortx_ens_33.private_ip} lookup 1",                  // Rule to Use Table 1
      "sudo ip rule add from all iif ens33 lookup 1",                                                     // Rule to Use Table 1 - All Incoming Interfaces
      "systemctl restart NetworkManager.service",
    ]
  }

  // Depends On...
  depends_on = [
    null_resource.init_terraform_user,
    aws_instance.cortx
  ]
}