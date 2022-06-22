//
// Configure a hosted Roue53 Zone to refer to resouces in the CORTX VPC
//

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_zone
resource "aws_route53_zone" "cortx" {

  // General
  name = "cortx.internal"

  // Internal - Define as an Internal Network
  vpc {
    vpc_id = aws_vpc.cortx.id
  }

  // Tags
  tags = {
    Environment = "Production"
    Owner       = "Terraform"
    Name        = "CORTX Internal"
  }
}

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record
resource "aws_route53_record" "mgmt" {

  // General
  zone_id = aws_route53_zone.cortx.zone_id
  name    = "mgmt"
  type    = "A"
  ttl     = "60"                            // Very Low TTL for Testing
  records = [aws_instance.cortx.private_ip] // Main CORTX Server Interface
}

// Resource: https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record
resource "aws_route53_record" "data" {

  // General
  zone_id = aws_route53_zone.cortx.zone_id
  name    = "data"
  type    = "A"
  ttl     = "60"                                            // Very Low TTL for Testing
  records = [aws_network_interface.cortx_ens_33.private_ip] // CORTX Server Secondary Interface

}