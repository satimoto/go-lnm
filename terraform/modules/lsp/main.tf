locals {
  availability_zone   = var.availability_zones[var.instance_number % length(var.availability_zones)]
  subnet_id           = var.private_subnet_ids[var.instance_number % length(var.private_subnet_ids)]
  lower_instance_name = lower(var.instance_name)
}

# -----------------------------------------------------------------------------
# Create the security group
# -----------------------------------------------------------------------------

resource "aws_security_group" "lsp_security_group" {
  name        = var.security_group_name
  description = "${var.instance_name} Security Group"
  vpc_id      = var.vpc_id

  tags = {
    Name = "${var.instance_name} Security Group"
  }
}

# -----------------------------------------------------------------------------
# Create the LSP security group rules
# -----------------------------------------------------------------------------

resource "aws_security_group_rule" "lsp_any_btc_p2p_ingress_rule" {
  type              = "ingress"
  from_port         = var.btc_p2p_port
  to_port           = var.btc_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "BTC-P2P from Any to ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_any_lnd_p2p_ingress_rule" {
  type              = "ingress"
  from_port         = var.lnd_p2p_port
  to_port           = var.lnd_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "LND-P2P from Any to ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_nlb_rest_ingress_rule" {
  type              = "ingress"
  from_port         = var.rest_port
  to_port           = var.rest_port
  protocol          = "tcp"
  cidr_blocks       = var.public_subnet_cidrs
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "REST from NLB to ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_private_rpc_ingress_rule" {
  type              = "ingress"
  from_port         = var.rpc_port
  to_port           = var.rpc_port
  protocol          = "tcp"
  cidr_blocks       = var.private_subnet_cidrs
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "RPC from Private to ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_any_btc_p2p_egress_rule" {
  type              = "egress"
  from_port         = var.btc_p2p_port
  to_port           = var.btc_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "BTC-P2P to Any from ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_any_lnd_p2p_egress_rule" {
  type              = "egress"
  from_port         = var.lnd_p2p_port
  to_port           = var.lnd_p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "LND-P2P to Any from ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_any_http_egress_rule" {
  type              = "egress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "HTTP to Any from ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_any_https_egress_rule" {
  type              = "egress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "HTTPS to Any from ${var.instance_name}"
}

resource "aws_security_group_rule" "lsp_private_rpc_egress_rule" {
  type              = "egress"
  from_port         = 50000
  to_port           = 50010
  protocol          = "tcp"
  cidr_blocks       = var.private_subnet_cidrs
  security_group_id = aws_security_group.lsp_security_group.id
  description       = "RPC to Private from ${var.instance_name}"
}

# -----------------------------------------------------------------------------
# Create the NAT security group rules
# -----------------------------------------------------------------------------

resource "aws_security_group_rule" "lsp_nat_ssh_ingress_rule" {
  type                     = "ingress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = var.nat_security_group_id
  security_group_id        = aws_security_group.lsp_security_group.id
  description              = "SSH from NAT to ${var.instance_name}"
}

resource "aws_security_group_rule" "nat_lsp_btc_p2p_ingress_rule" {
  type                     = "ingress"
  from_port                = var.btc_p2p_port
  to_port                  = var.btc_p2p_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.lsp_security_group.id
  security_group_id        = var.nat_security_group_id
  description              = "BTC-P2P from ${var.instance_name} to NAT"
}

resource "aws_security_group_rule" "nat_lsp_lnd_p2p_ingress_rule" {
  type                     = "ingress"
  from_port                = var.lnd_p2p_port
  to_port                  = var.lnd_p2p_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.lsp_security_group.id
  security_group_id        = var.nat_security_group_id
  description              = "LND-P2P from ${var.instance_name} to NAT"
}

resource "aws_security_group_rule" "nat_lsp_ssh_egress_rule" {
  type                     = "egress"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.lsp_security_group.id
  security_group_id        = var.nat_security_group_id
  description              = "SSH to ${var.instance_name} from NAT"
}

# -----------------------------------------------------------------------------
# Create the RDS security group rules
# -----------------------------------------------------------------------------

resource "aws_security_group_rule" "rds_lsp_pg_ingress_rule" {
  type                     = "ingress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.lsp_security_group.id
  security_group_id        = var.rds_security_group_id
  description              = "PG from ${var.instance_name} to RDS"
}

resource "aws_security_group_rule" "lsp_rds_pg_egress_rule" {
  type                     = "egress"
  from_port                = 5432
  to_port                  = 5432
  protocol                 = "tcp"
  source_security_group_id = var.rds_security_group_id
  security_group_id        = aws_security_group.lsp_security_group.id
  description              = "PG to RDS from ${var.instance_name}"
}

# -----------------------------------------------------------------------------
# Create the EC2 instance
# -----------------------------------------------------------------------------

resource "aws_instance" "lsp_instance" {
  ami                    = "ami-092f628832a8d22a5"
  availability_zone      = local.availability_zone
  instance_type          = var.ec2_instance_type
  key_name               = var.ec2_key_name
  subnet_id              = local.subnet_id
  vpc_security_group_ids = [aws_security_group.lsp_security_group.id]

  tags = {
    Name = var.instance_name
  }
}

# -----------------------------------------------------------------------------
# Create the EBS volume
# -----------------------------------------------------------------------------

resource "aws_ebs_volume" "lsp_ebs_volume" {
  availability_zone = local.availability_zone
  iops              = var.ebs_iops
  size              = var.ebs_size
  type              = var.ebs_type
  throughput        = var.ebs_throughput

  tags = {
    Name = var.instance_name
  }
}

resource "aws_volume_attachment" "lsp_ebs_volume_attachment" {
  device_name                    = var.ebs_attachment_device_name
  volume_id                      = aws_ebs_volume.lsp_ebs_volume.id
  instance_id                    = aws_instance.lsp_instance.id
  stop_instance_before_detaching = true
}

# -----------------------------------------------------------------------------
# Create the NLB
# -----------------------------------------------------------------------------

resource "aws_lb" "nlb" {
  name               = local.lower_instance_name
  internal           = false
  load_balancer_type = "network"
  subnets            = var.public_subnet_ids
}

# -----------------------------------------------------------------------------
# Create the NLB target groups
# -----------------------------------------------------------------------------

resource "aws_lb_target_group" "nlb_btc_p2p_target_group" {
  name     = "${local.lower_instance_name}-bitcoind"
  port     = var.btc_p2p_port
  protocol = "TCP"
  vpc_id   = var.vpc_id
  tags     = {}

  health_check {
    path     = "/health"
    port     = var.rest_port
    protocol = "HTTP"
  }
}

resource "aws_lb_target_group_attachment" "nlb_btc_p2p_target_group_attachment" {
  target_group_arn = aws_lb_target_group.nlb_btc_p2p_target_group.arn
  target_id        = aws_instance.lsp_instance.id
  port             = var.btc_p2p_port
}

resource "aws_lb_target_group" "nlb_lnd_p2p_target_group" {
  name     = "${local.lower_instance_name}-lnd"
  port     = var.lnd_p2p_port
  protocol = "TCP"
  vpc_id   = var.vpc_id
  tags     = {}

  health_check {
    path     = "/health"
    port     = var.rest_port
    protocol = "HTTP"
  }
}

resource "aws_lb_target_group_attachment" "nlb_lnd_p2p_target_group_attachment" {
  target_group_arn = aws_lb_target_group.nlb_lnd_p2p_target_group.arn
  target_id        = aws_instance.lsp_instance.id
  port             = var.lnd_p2p_port
}

# -----------------------------------------------------------------------------
# Create the NLB listeners
# -----------------------------------------------------------------------------

resource "aws_lb_listener" "nlb_btc_p2p_listener" {
  load_balancer_arn = aws_lb.nlb.arn
  port              = var.btc_p2p_port
  protocol          = "TCP"
  tags              = {}

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.nlb_btc_p2p_target_group.arn
  }
}

resource "aws_lb_listener" "nlb_lnd_p2p_listener" {
  load_balancer_arn = aws_lb.nlb.arn
  port              = var.lnd_p2p_port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.nlb_lnd_p2p_target_group.arn
  }
}

# -----------------------------------------------------------------------------
# Create the Route53 record to the NLB
# -----------------------------------------------------------------------------

resource "aws_route53_record" "service" {
  zone_id = var.route53_zone_id
  name    = var.route53_domain_name
  type    = "A"

  alias {
    name                   = aws_lb.nlb.dns_name
    zone_id                = aws_lb.nlb.zone_id
    evaluate_target_health = false
  }
}
