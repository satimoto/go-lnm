## Install Rebalance LND

Clone the rebalance-lnd repository
```bash
git clone https://github.com/satimoto/rebalance-lnd.git
cd rebalance-lnd/
```
Install pip python package manager
```bash
sudo apt install python3-pip
```
Install rebalance-lnd dependencies
```bash
pip install -r requirements.txt
pip install testresources
```

## Install Charge LND

Create a macaroon for charge-lnd
```bash
lncli bakemacaroon offchain:read offchain:write onchain:read info:read --save_to=~/.lnd/data/chain/bitcoin/mainnet/charge-lnd.macaroon
```
Clone the charge-lnd repository
```bash
git clone https://github.com/satimoto/charge-lnd.git
cd charge-lnd/
```
Install rebalance-lnd dependencies
```bash
pip install -r requirements.txt .
```
Make the config directory
```bash
mkdir ~/.charge-lnd
```
Edit the config file `~/.charge-lnd/charge-lnd.config`
```bash
[default]
# 'default' is special, it is used if no other policy matches a channel
strategy = static
base_fee_msat = 1_000
fee_ppm = 10

[mydefaults]
# no strategy, so this only sets some defaults
base_fee_msat = 1_000
min_fee_ppm_delta = 5

[zerobasefee-initiative]
# set base_fee_msat default to 0 if peer sets zero base fees
chan.max_base_fee_msat = 0
base_fee_msat = 0

[ignored_channels]
# don't let charge-lnd set fees (strategy=ignore) for channels to/from the specified nodes
node.id = 02da8d5a759ee9e4438da617cfdb61c87f723fb76c4b6371b877d0347abe953a4f,
  0266ad254117f16f16c3457e081e6207e91c5e414477a208cf4d9c633322799038
strategy = ignore

[known-one-way-drains]
node.id = 033d8656219478701227199cbd6f670335c8d408a92ae88b962c49d4dc0e83e025,
  03cde60a6323f7122d5178255766e38114b4722ede08f7c9e0c5df9b912cc201d6
  021c97a90a411ff2b10dc2a8e32de2f29d2fa49d41bfbb52bd416e460db0747d0d,
  03037dc08e9ac63b82581f79b662a4d0ceca8a8ca162b1af3551595b8f2d97b70a,
  030c3f19d742ca294a55c00376b3b355c3c90d61c6b6b39554dbc7ac19b141c14f,
  03d607f3e69fd032524a867b288216bfab263b6eaee4e07783799a6fe69bb84fac,
  02a04446caa81636d60d63b066f2814cbd3a6b5c258e3172cbdded7a16e2cfff4c
strategy = static
fee_ppm = 3500

[expensive]
# match channels where the peer node has set a high (>=8_000 ppm) fee rate
# and set the same fee rate on our side (strategy=match_peer)
chan.min_fee_ppm = 8_000
strategy = match_peer

[private-node]
# charge non-routing (private=true) peers a bit more for our service
chan.private = true
strategy = static
fee_ppm = 10

[beginner-node]
# set lower fees on channels with smaller peers, that have few channels (4-8 channels) total
# and limited node size (max_capacity=1_000_000)
node.max_capacity = 1_000_000
node.min_channels = 4
node.max_channels = 8
strategy = static
base_fee_msat = 0
fee_ppm = 1

[encourage-routing]
# 'autobalance' (lower fees so using outbound is more attractive) larger channels (min_capacity 2M sats)
# to larger nodes (node has at least 50M sats) if balance ratio >= 0.9 (more than 90% on our side)
chan.min_ratio = 0.9
chan.min_capacity = 2_000_000
node.min_capacity = 5_000_0000
strategy = static
base_fee_msat = 500
fee_ppm = 5

[discourage-routing]
# 'autobalance' (higher fees so using outbound is less attractive) larger channels (min_capacity 2M sats)
# to larger nodes (node has at least 50M sats) if balance ratio <= 0.2 (less than 20% on our side)
chan.max_ratio = 0.2
chan.min_capacity = 2_000_000
node.min_capacity = 5_000_0000
strategy = static
base_fee_msat = 10_000
fee_ppm = 500

[proportional-routing]
# 'proportional' can also be used to auto balance (lower fee rate when low remote balance & higher rate when higher remote balance)
# fee_ppm decreases linearly with the channel balance ratio (min_fee_ppm when ratio is 1, max_fee_ppm when ratio is 0)
chan.min_ratio = 0.2
chan.max_ratio = 0.9
chan.min_capacity = 1_000_000
strategy = proportional
min_fee_ppm = 10
max_fee_ppm = 200

[recover-cost]
# use cost strategy if we are the initiator, using a base fee of 1 sat, and a cost_factor of 1.5 to recover 150% of costs when
# channel is depleted to 100% outgoing
chan.initiator = true
strategy = cost
cost_factor = 1.5
```
Add a `crontab` entry for charge-lnd
```bash
crontab -e
```
```bash
*/5 * * * * /home/ubuntu/.local/bin/charge-lnd -c /home/ubuntu/.charge-lnd/charge-lnd.config | systemd-cat -t charge-lnd
```
Edit `~/.profile` and add the lines
```bash
alias loglsp="journalctl -ft lsp"
alias logall="journalctl -ft bitcoind -t lnd -t lsp -t chargelnd"
```