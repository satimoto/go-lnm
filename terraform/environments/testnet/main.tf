provider "aws" {
  region                = var.region
  forbidden_account_ids = var.forbidden_account_ids
  profile               = "satimoto-testnet"
}

provider "aws" {
  alias                 = "us_east_1"
  region                = "us-east-1"
  forbidden_account_ids = var.forbidden_account_ids
  profile               = "satimoto-testnet"
}

provider "aws" {
  alias                 = "zone_owner"
  region                = var.region
  forbidden_account_ids = var.forbidden_account_ids
  profile               = "satimoto-common"
}

# -----------------------------------------------------------------------------
# Backends
# -----------------------------------------------------------------------------

data "terraform_remote_state" "infrastructure" {
  backend = "s3"

  config = {
    bucket  = "satimoto-terraform-testnet"
    key     = "infrastructure.tfstate"
    region  = "eu-central-1"
    profile = "satimoto-testnet"
  }
}

terraform {
  backend "s3" {
    bucket  = "satimoto-terraform-testnet"
    key     = "lsp.tfstate"
    region  = "eu-central-1"
    profile = "satimoto-testnet"
  }
}

# -----------------------------------------------------------------------------
# Modules
# -----------------------------------------------------------------------------

resource "aws_security_group_rule" "nat_any_btc_p2p_egress_rule" {
  type              = "egress"
  from_port         = var.btc_p2p_port
  to_port           = var.btc_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = data.terraform_remote_state.infrastructure.outputs.nat_security_group_id
  description       = "BTC-P2P to Any from NAT"
}

resource "aws_security_group_rule" "nat_any_lnd_p2p_egress_rule" {
  type              = "egress"
  from_port         = var.lnd_p2p_port
  to_port           = var.lnd_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = data.terraform_remote_state.infrastructure.outputs.nat_security_group_id
  description       = "LND-P2P to Any from NAT"
}

module "lsp" {
  count              = var.lsp_count
  source             = "../../modules/lsp"
  availability_zones = var.availability_zones
  region             = var.region

  vpc_id                = data.terraform_remote_state.infrastructure.outputs.vpc_id
  private_subnet_ids    = data.terraform_remote_state.infrastructure.outputs.private_subnet_ids
  public_subnet_cidrs   = data.terraform_remote_state.infrastructure.outputs.public_subnet_cidrs
  public_subnet_ids     = data.terraform_remote_state.infrastructure.outputs.public_subnet_ids
  security_group_name   = "lsp${count.index + 1}-instance"
  ecs_security_group_id = data.terraform_remote_state.infrastructure.outputs.ecs_security_group_id
  nat_security_group_id = data.terraform_remote_state.infrastructure.outputs.nat_security_group_id
  rds_security_group_id = data.terraform_remote_state.infrastructure.outputs.rds_security_group_id
  instance_number       = count.index
  instance_name         = "LSP${count.index + 1}"
  ec2_instance_type     = var.ec2_instance_type
  ec2_key_name          = "lsp${count.index + 1}"
  ebs_iops              = 3000
  ebs_size              = 30
  ebs_type              = "gp3"
  ebs_throughput        = 125
  route53_zone_id       = data.terraform_remote_state.infrastructure.outputs.route53_zone_id
  route53_domain_name   = "lsp${count.index + 1}.${data.terraform_remote_state.infrastructure.outputs.route53_zone_name}"
  rest_port             = var.rest_port
  rpc_port              = var.rpc_port
  btc_p2p_port          = var.btc_p2p_port
  lnd_p2p_port          = var.lnd_p2p_port
}
