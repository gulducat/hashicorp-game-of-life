# TODO: why need this?
terraform {
  required_version = ">= 0.13"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "2.70.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

module "nomad-oss" {
  source                = "./terraform-aws-nomad-starter/modules/nomad_cluster"
  #                         db                  omar               kingsley
  allowed_inbound_cidrs = ["136.49.27.204/32", "69.237.40.71/32", "209.6.27.148/32"]
  vpc_id                = "vpc-bc011fc6" # default vpc in product_delivery_prod
  consul_version        = "1.8.5"
  nomad_version         = "0.12.8"
  owner                 = "dbennett"
  name_prefix           = "dbennett-"
  key_name              = "dbennett-test"
  nomad_servers         = 1
  nomad_clients         = 5
  instance_type         = "m5.xlarge" # defualt = m5.large
}
