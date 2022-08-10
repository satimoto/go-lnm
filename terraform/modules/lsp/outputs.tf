
output "lsp_security_group_id" {
  description = "The security group ID of the LSP instance"
  value       = aws_security_group.lsp_security_group.id
}

output "lsp_private_ip" {
  description = "The private IP of the LSP instance"
  value       = aws_instance.lsp_instance.private_ip
}

output "lsp_ebs_id" {
  description = "The volume ID of the LSP EBS volume"
  value       = aws_ebs_volume.lsp_ebs_volume.id
}

output "lsp_ebs_arn" {
  description = "The volume ARN of the LSP EBS volume"
  value       = aws_ebs_volume.lsp_ebs_volume.arn
}

