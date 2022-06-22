// Deployer Data

//
// Resource: https://registry.terraform.io/providers/hashicorp/http/latest/docs/data-sources/http
data "http" "deployerip" {
  url = "http://ipv4.icanhazip.com"
}

