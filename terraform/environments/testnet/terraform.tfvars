region = "eu-central-1"

availability_zones = ["eu-central-1a", "eu-central-1b", "eu-central-1c"]

deployment_stage = "testnet"

forbidden_account_ids = ["490833747373"]

# -----------------------------------------------------------------------------
# Module service-api
# -----------------------------------------------------------------------------

lsp_count = 1

ec2_instance_type = "t3.medium"

metric_port = 9102

rest_port = 9002

rpc_port = 50000

btc_p2p_port = 18333

lnd_p2p_port = 9735
