variable "region" {
  description = "The AWS region"
  default     = "eu-central-1"
}

variable "availability_zones" {
  description = "A list of Availability Zones where subnets and DB instances can be created"
}

variable "deployment_stage" {
  description = "The deployment stage"
  default     = "testnet"
}

variable "forbidden_account_ids" {
  description = "The forbidden account IDs"
  type        = list(string)
  default     = []
}

# -----------------------------------------------------------------------------
# Module lsp
# -----------------------------------------------------------------------------

variable "lsp_count" {
  description = "Number of LSPs to provision"
}

variable "ec2_instance_type" {
  description = "The instance type to use for the EC2 instance"
}

variable "rest_port" {
  description = "The port to access metrics and health"
  default     = 9002
}

variable "rpc_port" {
  description = "The port to access RPC server"
  default     = 50000
}

variable "btc_p2p_port" {
  description = "The port bitcoind uses for p2p connections"
  default     = 8333
}

variable "lnd_p2p_port" {
  description = "The port lnd uses for p2p connections"
  default     = 9735
}