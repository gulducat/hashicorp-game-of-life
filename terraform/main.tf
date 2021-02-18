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
  allowed_inbound_cidrs = ["136.49.27.204/32"]
  vpc_id                = "vpc-3e140a44" # default vpc in engserv_sandbox_dev
  consul_version        = "1.9.2"
  nomad_version         = "1.0.3-2"
  owner                 = "dbennett"
  name_prefix           = "dbennett"
  key_name              = "dbennett-test"
  nomad_servers         = 1
  nomad_clients         = 7
  instance_type         = "c5.2xlarge" # 8cpu 16m
}
