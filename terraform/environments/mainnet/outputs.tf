
output "lsp_security_group_ids" {
  description = "The security group IDs of the LSP instances"
  value       = module.lsp.*.lsp_security_group_id
}

output "lsp_private_ips" {
  description = "The private IPs of the LSP instances"
  value       = module.lsp.*.lsp_private_ip
}
