provider "aws" {
  region                = var.region
  forbidden_account_ids = var.forbidden_account_ids
  profile               = "satimoto-mainnet"
}

provider "aws" {
  alias                 = "us_east_1"
  region                = "us-east-1"
  forbidden_account_ids = var.forbidden_account_ids
  profile               = "satimoto-mainnet"
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
    bucket  = "satimoto-terraform-mainnet"
    key     = "infrastructure.tfstate"
    region  = "eu-central-1"
    profile = "satimoto-mainnet"
  }
}

terraform {
  backend "s3" {
    bucket  = "satimoto-terraform-mainnet"
    key     = "lnm.tfstate"
    region  = "eu-central-1"
    profile = "satimoto-mainnet"
  }
}

# -----------------------------------------------------------------------------
# Create the backup bucket
# -----------------------------------------------------------------------------

resource "aws_s3_bucket" "backup_bucket" {
  bucket = "satimoto-${var.service_name}-${var.deployment_stage}-channel-backup"
}

resource "aws_s3_bucket_acl" "backup_bucket_acl" {
  bucket = aws_s3_bucket.backup_bucket.id
  acl    = "private"
}

resource "aws_s3_bucket_public_access_block" "backup_bucket_public_access_block" {
  bucket = aws_s3_bucket.backup_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_iam_user" "backup_user" {
  name = "${var.service_name}-backup-user"
}

resource "aws_iam_policy" "backup_s3_user_policy" {
  name = "${var.service_name}-backup-s3-user-policy"

  policy = templatefile("../../resources/backup-user-policy.json", {
    bucket_name = "satimoto-${var.service_name}-${var.deployment_stage}-channel-backup"
  })
}

resource "aws_iam_user_policy_attachment" "backup_s3_user_policy_attachment" {
  user       = aws_iam_user.backup_user.name
  policy_arn = aws_iam_policy.backup_s3_user_policy.arn
}

resource "aws_iam_access_key" "backup_saccess_key" {
  user = aws_iam_user.backup_user.name
}

# -----------------------------------------------------------------------------
# Modules
# -----------------------------------------------------------------------------

module "subdomain_zone" {
  providers = {
    aws.zone_owner = aws.zone_owner
  }
  source             = "git::https://github.com/satimoto/terraform-infrastructure.git//modules/subdomain-zone?ref=develop"
  availability_zones = var.availability_zones
  region             = var.region

  domain_name     = data.terraform_remote_state.infrastructure.outputs.route53_zone_name
  subdomain_name  = var.subdomain_name
  route53_zone_id = data.terraform_remote_state.infrastructure.outputs.route53_zone_id
}

data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "satimoto_db_password" {
  name = var.rds_satimoto_db_password_ssm_key
}

data "aws_ssm_parameter" "lnd_macaroon" {
  name = var.lnm_lnd_macaroon_ssm_key
}

module "service-lnm" {
  source             = "git::https://github.com/satimoto/terraform-infrastructure.git//modules/service?ref=f9cad99f17c1d7c14273b9433e249922a2b92544"
  availability_zones = var.availability_zones
  deployment_stage   = var.deployment_stage
  region             = var.region

  vpc_id                         = data.terraform_remote_state.infrastructure.outputs.vpc_id
  private_subnet_ids             = data.terraform_remote_state.infrastructure.outputs.private_subnet_ids
  route53_zone_id                = module.subdomain_zone.route53_zone_id
  alb_security_group_id          = data.terraform_remote_state.infrastructure.outputs.alb_security_group_id
  alb_dns_name                   = data.terraform_remote_state.infrastructure.outputs.alb_dns_name
  alb_zone_id                    = data.terraform_remote_state.infrastructure.outputs.alb_zone_id
  alb_listener_arn               = data.terraform_remote_state.infrastructure.outputs.alb_listener_arn
  ecs_cluster_id                 = data.terraform_remote_state.infrastructure.outputs.ecs_cluster_id
  ecs_security_group_id          = data.terraform_remote_state.infrastructure.outputs.ecs_security_group_id
  ecs_task_execution_role_arn    = data.terraform_remote_state.infrastructure.outputs.ecs_task_execution_role_arn
  service_name                   = var.service_name
  service_domain_names           = ["${var.subdomain_name}.${data.terraform_remote_state.infrastructure.outputs.route53_zone_name}"]
  service_desired_count          = var.service_desired_count
  service_container_name         = var.service_name
  service_container_port         = var.service_container_port
  task_network_mode              = var.task_network_mode
  task_cpu                       = var.task_cpu
  task_memory                    = var.task_memory
  target_health_path             = var.target_health_path
  target_health_interval         = var.target_health_interval
  target_health_timeout          = var.target_health_timeout
  target_health_matcher          = var.target_health_matcher
  service_discovery_namespace_id = data.terraform_remote_state.infrastructure.outputs.ecs_service_discovery_namespace_id

  task_container_definitions = templatefile("../../resources/task-container-definitions.json", {
    account_id                       = data.aws_caller_identity.current.account_id
    image_tag                        = "mainnet"
    region                           = var.region
    service_name                     = var.service_name
    service_container_port           = var.service_container_port
    service_metric_port              = var.service_metric_port
    rpc_container_port               = var.env_rpc_port
    task_network_mode                = var.task_network_mode
    env_accounting_currency          = var.env_accounting_currency
    env_backup_aws_region            = var.region
    env_backup_aws_access_key_id     = aws_iam_access_key.backup_saccess_key.id
    env_backup_aws_secret_access_key = aws_iam_access_key.backup_saccess_key.secret
    env_backup_aws_s3_bucket         = aws_s3_bucket.backup_bucket.id
    env_circuit_percent              = var.env_circuit_percent
    env_db_user                      = "satimoto"
    env_db_pass                      = data.aws_ssm_parameter.satimoto_db_password.value
    env_db_host                      = "${data.terraform_remote_state.infrastructure.outputs.rds_cluster_endpoint}:${data.terraform_remote_state.infrastructure.outputs.rds_cluster_port}"
    env_db_name                      = "satimoto"
    env_default_tax_percent          = var.env_default_tax_percent
    env_fcm_api_key                  = var.env_fcm_api_key
    env_ferp_rpc_address             = "ferp.${data.terraform_remote_state.infrastructure.outputs.ecs_service_discovery_namespace_name}:${var.env_ferp_rpc_port}"
    env_lnd_tls_cert                 = var.env_lnd_tls_cert
    env_lnd_grpc_host                = var.env_lnd_grpc_host
    env_lnd_macaroon                 = data.aws_ssm_parameter.lnd_macaroon.value
    env_ocpi_rpc_address             = "ocpi.${data.terraform_remote_state.infrastructure.outputs.ecs_service_discovery_namespace_name}:${var.env_ocpi_rpc_port}"
    env_metric_port                  = var.service_metric_port
    env_rest_port                    = var.service_container_port
    env_rpc_host                     = "${var.subdomain_name}.${data.terraform_remote_state.infrastructure.outputs.ecs_service_discovery_namespace_name}"
    env_rpc_port                     = var.env_rpc_port
    env_shutdown_timeout             = var.env_shutdown_timeout
    env_update_unsettled_invoices    = var.env_update_unsettled_invoices
  })
}
