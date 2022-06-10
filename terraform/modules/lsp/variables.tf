variable "region" {
  description = "The AWS region"
  default     = "eu-central-1"
}

variable "deployment_stage" {
  description = "The deployment stage"
  default     = "testnet"
}

variable "availability_zones" {
  description = "A list of Availability Zones where subnets and DB instances can be created"
}

# -----------------------------------------------------------------------------
# Module lsp
# -----------------------------------------------------------------------------

variable "vpc_id" {
  description = "The ID of the VPC"
}

variable "private_subnet_cidrs" {
  description = "The CIDRs of private subnets"
}

variable "private_subnet_ids" {
  description = "The IDs of private subnets"
}

variable "public_subnet_cidrs" {
  description = "The CIDRs of public subnets"
}

variable "public_subnet_ids" {
  description = "The IDs of public subnets"
}

variable "security_group_name" {
  description = "The EC2 security group name"
}

variable "nat_security_group_id" {
  description = "The security group ID of the NAT"
}

variable "rds_security_group_id" {
  description = "The security group ID of the RDS"
}

variable "instance_number" {
  description = "The LSP instance number"
}

variable "instance_name" {
  description = "The EC2 instance name"
}

variable "ec2_instance_type" {
  description = "The instance type to use for the EC2 instance"
}

variable "ec2_key_name" {
  description = "Key name of the Key Pair to use for the EC2 instance"
}

variable "ebs_iops" {
  description = "The amount of IOPS to provision for the disk"
}

variable "ebs_size" {
  description = "The size of the drive in GiBs"
}

variable "ebs_type" {
  description = "The type of EBS volume"
}

variable "ebs_throughput" {
  description = "The throughput that the volume supports, in MiB/s"
}

variable "ebs_attachment_device_name" {
  description = "The device name to expose to the instance"
  default     = "/dev/sdh"
}

variable "route53_zone_id" {
  description = "The Route53 Zone ID"
}

variable "route53_domain_name" {
  description = "The Route53 full domain name"
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
