Go through the following steps each time you need to deploy a new LSP

## Key Pair
Add a new Key Pair matching the next lsp to be deployed e.g. `lsp1`

Save PEM file to your local .ssh directory

## Terraform
Run terraform apply
```bash
terraform init
terraform apply
```

## SSH Key
Change permissions to restrict access
```bash
chmod 0600 lspX.pem
```
Add the pem identity to the ssh-agent
```bash
ssh-add -K lspX.pem
```
SSH into the NAT
```bash
ssh -A satimoto.nat
```
Add LSP into hosts, edit `/etc/hosts` and add a line:
```bash
XXX.XXX.XXX.XXX satimoto.lspX
```
SSH into the LSP
```bash
ssh ubuntu@satimoto.lspX
```

## Update installed packages
```bash
sudo apt update && sudo apt upgrade -y
```

## Increase file descriptor limit
Edit `/etc/sysctl.conf` and add this line:
```bash
fs.file-max=512000
```
Then reboot
```bash
sudo reboot
```

## Setup EBS disk
List storage
```bash
lsblk
```
which should show a volume named `nvme0n1p1`
```bash
NAME          MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
nvme1n1       259:0    0  30G  0 disk
nvme0n1       259:1    0  30G  0 disk
├─nvme0n1p1   259:2    0  30G  0 part /
└─nvme0n1p128 259:3    0   1M  0 part
```
Check the volume is empty
```bash
sudo file -s /dev/nvme1n1
```
when empty should show
```bash
/dev/nvme1n1: data
```
Format the disk as ext4
```bash
sudo mkfs -t ext4 /dev/nvme1n1
```
Make a directory for the volume and mount it
```bash
sudo mkdir /blockchain
sudo mount /dev/nvme1n1 /blockchain/
cd /blockchain
```
Check the volume size
```bash
df -h .
```
```bash
Filesystem      Size  Used Avail Use% Mounted on
/dev/nvme1n1     30G   45M   28G   1% /blockchain
```
Automatically mount the volume, but first backup the config
```bash
sudo cp /etc/fstab /etc/fstab.bak
```
Edit `/etc/fstab` and add this line:
```bash
/dev/nvme1n1                                  /blockchain ext4   defaults,nofail   0   0
```
Mount the volume
```bash
sudo mount -a
```
Change ownership of the directory
```bash
sudo chown `whoami` /blockchain
```

## Setup firewall
```
Enable ufw
```bash
sudo ufw logging on
sudo ufw enable
```
Allow ports
```bash
sudo ufw allow OpenSSH
sudo ufw allow 9002
sudo ufw allow 9102
sudo ufw allow 9735
sudo ufw allow 10009
sudo ufw allow 50000
# Mainnet
sudo ufw allow 8333
# Testnet
sudo ufw allow 18333
```

## Setup flood protection
```bash
sudo iptables -N syn_flood
sudo iptables -A INPUT -p tcp --syn -j syn_flood
sudo iptables -A syn_flood -m limit --limit 1/s --limit-burst 3 -j RETURN
sudo iptables -A syn_flood -j DROP
sudo iptables -A INPUT -p icmp -m limit --limit 1/s --limit-burst 1 -j ACCEPT
sudo iptables -A INPUT -p icmp -m limit --limit 1/s --limit-burst 1 -j LOG --log-prefix PING-DROP:
sudo iptables -A INPUT -p icmp -j DROP
sudo iptables -A OUTPUT -p icmp -j ACCEPT
```

## Install Bitcoin Core
Install compilation dependencies
```bash
sudo apt install git build-essential libtool autotools-dev automake pkg-config libssl-dev libevent-dev bsdmainutils libboost-system-dev libboost-filesystem-dev libboost-chrono-dev libboost-program-options-dev libboost-test-dev libboost-thread-dev libminiupnpc-dev libzmq3-dev
```
Clone the bitcoin repository
```bash
git clone -b v23.0 https://github.com/bitcoin/bitcoin.git
cd bitcoin/
```
Configure build
```bash
./autogen.sh
./configure CXXFLAGS="--param ggc-min-expand=1 --param ggc-min-heapsize=32768" --enable-cxx --with-zmq --without-gui --disable-shared --with-pic --disable-tests --disable-bench --enable-upnp-default --disable-wallet
```
Make and install
```bash
make -j "$(($(nproc)+1))"
sudo make install
```
Create bitcoin data and config directories
```bash
mkdir -p /blockchain/.bitcoin/data
mkdir ~/.bitcoin
```
Download and run rpc auth script, these will be added to `bitcoin.conf` and `lnd.conf`
```bash
wget https://raw.githubusercontent.com/bitcoin/bitcoin/master/share/rpcauth/rpcauth.py
python3 ./rpcauth.py bitcoinrpc
```
Edit the `~/.bitcoin/bitcoin.conf` file, use `getbestblockhash` to get the current chain tip hash
```bash
# Set the best block hash here:
assumevalid=

# Set the data directory to the storage directory
datadir=/blockchain/.bitcoin/data

# Set the number of megabytes of RAM to use, set to like 50% of available memory
dbcache=2000

# Add visibility into mempool and RPC calls for potential LND debugging
debug=mempool
debug=rpc

# Turn off the wallet, it won't be used
disablewallet=1

# Turn on listening mode
listen=1

# Constrain the mempool to the number of megabytes needed:
maxmempool=300

# Limit uploading to peers
maxuploadtarget=1000

# Turn off serving SPV nodes
nopeerbloomfilters=1
peerbloomfilters=0

# Don't accept deprecated multi-sig style
permitbaremultisig=0

# Set the RPC auth to what was set above
rpcauth=

# Turn on the RPC server
server=1

# Print log to console
printtoconsole=1

# Set testnet if needed
testnet=1

# Turn on transaction pruning
prune=16384

# Turn on ZMQ publishing
zmqpubrawblock=tcp://127.0.0.1:28332
zmqpubrawtx=tcp://127.0.0.1:28333
```
Add a systemd service to manage bitcoind. Edit `/etc/systemd/system/bitcoind.service`
```bash
[Unit]
Description=Bitcoin daemon
After=network-online.target
Wants=network-online.target

[Service]
User=ubuntu
Group=ubuntu
Type=forking
ExecStart=/usr/local/bin/bitcoind -daemonwait \
                                  -pid=/home/ubuntu/.bitcoin/bitcoind.pid \
                                  -conf=/home/ubuntu/.bitcoin/bitcoin.conf \
                                  -datadir=/blockchain/.bitcoin/data
PIDFile=/home/ubuntu/.bitcoin/bitcoind.pid
KillMode=process
Restart=on-failure
SyslogIdentifier=bitcoind
TimeoutStartSec=infinity
TimeoutStopSec=600

[Install]
WantedBy=multi-user.target
```
Start bitcoind and enable running on startup
```bash
sudo systemctl start bitcoind
sudo systemctl enable bitcoind
```
Edit `~/.profile` and add the line
```bash
alias logbitcoind="journalctl -fu bitcoind"
```
If IBD is failing or bitcoind stops, check the kernal logs
```bash
tail -n 100 /var/log/kern.log
```

## Install Go
Download Go
```bash
cd 
wget https://golang.org/dl/go1.17.5.linux-amd64.tar.gz
```
Extract it
```bash
sudo tar -xvf go1.17.5.linux-amd64.tar.gz
```
Install it and remove the download
```bash
sudo mv go /usr/local && rm go1.17.5.linux-amd64.tar.gz
```
Make a directory for it
```bash
mkdir ~/go
```
Edit `~/.profile` and add the GOPATH and alias if testnet
```bash
GOPATH=$HOME/go
PATH="$HOME/bin:$GOPATH/bin:$HOME/.local/bin:/usr/local/go/bin:$PATH"
# Testnet
alias lncli="lncli --network=testnet"
```

## Create LND database user
Generate a new password for a database user. 
Create the user login role with name (e.g. `lnd1`) and generated password.  
Create the database (e.g. `lnd1`) with created login role.
Keep credentials to be added to `lnd.conf`.

## Install LND
```
Clone the lnd repository
```bash
cd ~/
git clone https://github.com/satimoto/lnd.git
cd lnd/
```
Checkout branch
```bash
git checkout -b allocate-alias origin/allocate-alias
```
Make lnd
```bash
make && make install tags="autopilotrpc chainrpc devrpc neutrinorpc invoicesrpc peersrpc routerrpc signrpc verrpc walletrpc watchtowerrpc wtclientrpc monitoring kvdb_postgres"
```
Create lnd data and config directories
```bash
mkdir ~/.lnd
```
Edit the `~/.lnd/lnd.conf` file
```bash
[Application Options]
# Allow push payments
accept-keysend=1

# Public network name
alias=lsp1.satimoto.com

# Allow gift routes
allow-circular-route=1

# Public hex color
color=#9911FF

# Reduce the cooperative close chain fee
coop-close-target-confs=1000

# Log levels
debuglevel=CNCT=debug,CRTR=debug,HSWC=debug,NTFN=debug,RPCS=debug

# Public P2P IP (remove this if using Tor)
externalip=lsp1.satimoto.com

# Mark unpayable, unpaid invoices as deleted
gc-canceled-invoices-on-startup=1
gc-canceled-invoices-on-the-fly=1

# Avoid historical graph data sync
ignore-historical-gossip-filters=1

# Listen (not using Tor? Remove this)
# listen=localhost

# Set the maximum amount of commit fees in a channel
max-channel-fee-allocation=1.0

# Set the max timeout blocks of a payment
max-cltv-expiry=5000

# Allow commitment fee to rise on anchor channels
max-commit-fee-rate-anchors=100

# Pending channel limit
maxpendingchannels=10

# Min inbound channel limit
minchansize=5000000

# gRPC socket binding
rpclisten=0.0.0.0:10009

# Avoid slow startup time
sync-freelist=1

# Avoid high startup overhead
stagger-initial-reconnect=1

# Delete and recreate RPC TLS certificate when details change or cert expires
tlsautorefresh=1

# Do not include IPs in the RPC TLS certificate
tlsdisableautofill=1

# Add DNS to the RPC TLS certificate
tlsextradomain=lsp1.satimoto.com

# The full path to a file (or pipe/device) that contains the password for unlocking the wallet
# wallet-unlock-password-file=/home/ubuntu/.lnd/wallet_password

# Reject requests with non-zero push amounts
rejectpush=true

# Hold HTLCs until there is an interceptor
requireinterceptor=true

[Bitcoin]
# Turn on Bitcoin mode
bitcoin.active=1

# Set the channel confs to wait for channels
bitcoin.defaultchanconfs=2

# Forward base fee
bitcoin.basefee=0

# Forward fee rate in parts per million
bitcoin.feerate=1000

# Set bitcoin.testnet=1 or bitcoin.mainnet=1 as appropriate
bitcoin.testnet=1

# Set the lower bound for HTLCs
bitcoin.minhtlc=1

# Set backing node, bitcoin.node=neutrino or bitcoin.node=bitcoind
bitcoin.node=bitcoind

# Set CLTV forwarding delta time
bitcoin.timelockdelta=40

[bitcoind]
# Configuration for using Bitcoin Core backend

# Set the password to what the auth script said
bitcoind.rpcpass=

# Set the username
bitcoind.rpcuser=bitcoinrpc

# Set the ZMQ listeners
bitcoind.zmqpubrawblock=tcp://127.0.0.1:28332
bitcoind.zmqpubrawtx=tcp://127.0.0.1:28333

[db]
# Set the database backend
db.backend=postgres

[postgres]
# Set the postgres database connection string
db.postgres.dsn=postgresql://lnd1:dbpass@satimoto.cluster-csvwlfckqqfq.eu-central-1.rds.amazonaws.com:5432/lnd1

# Set the postgres database connection timeout
db.postgres.timeout=0

[protocol]
# Enable SCID alias support
protocol.option-scid-alias=true

# Enable zero-conf support
protocol.zero-conf=true

# Enable large channels support
protocol.wumbo-channels=1

[routerrpc]
# Set default chance of a hop success
routerrpc.apriorihopprob=0.5

# Start to ignore nodes if they return many failures (set to 1 to turn off)
routerrpc.aprioriweight=0.75

# Set minimum desired savings of trying a cheaper path
routerrpc.attemptcost=10
routerrpc.attemptcostppm=10

# Set the number of historical routing records
routerrpc.maxmchistory=10000

# Set the min confidence in a path worth trying
routerrpc.minrtprob=0.005

# Set the time to forget past routing failures
routerrpc.penaltyhalflife=6h0m0s

[routing]
# Remove channels from graph that have one side that hasn't made announcements
routing.strictgraphpruning=1
```
Add a systemd service to manage LND. Edit `/etc/systemd/system/lnd.service`
```bash
[Unit]
Description=LND daemon
After=bitcoind.service
Wants=bitcoind.service

[Service]
User=ubuntu
Group=ubuntu
ExecStart=/home/ubuntu/go/bin/lnd --configfile=/home/ubuntu/.lnd/lnd.conf
ExecStop=/home/ubuntu/go/bin/lncli stop
Restart=on-failure
RestartSec=30

[Install]
WantedBy=multi-user.target
```
**Wait until Bitcoin Core has finished the IBD**

Start lnd and enable running on startup
```bash
sudo systemctl start lnd
sudo systemctl enable lnd
```
Create the wallet password
```bash
openssl rand -hex 21 > ~/.lnd/wallet_password
cat ~/.lnd/wallet_password
```
Create the wallet, using above password and no cipher seed password
```bash
lncli create
```
Edit the `~/.lnd/lnd.conf` file, uncommenting the line
```bash
wallet-unlock-password-file=/home/ubuntu/.lnd/wallet_password
```
Edit `~/.profile` and add the line
```bash
alias loglnd="journalctl -fu lnd"
```

## Install LSP
Edit file `~/.gitconfig`
```bash
[credential]
        helper = store
[credential "https://github.com"]
        useHttpPath = true
```
Edit file `~/.git-credentials`
```bash
https://<PERSONAL_ACCESS_TOKEN>:@github.com/satimoto/go-datastore
https://<PERSONAL_ACCESS_TOKEN>:@github.com/satimoto/go-ferp
https://<PERSONAL_ACCESS_TOKEN>:@github.com/satimoto/go-lnm
https://<PERSONAL_ACCESS_TOKEN>:@github.com/satimoto/go-ocpi
```
Clone the lnd repository
```bash
cd ~/
git clone https://github.com/satimoto/go-lnm
cd go-lsp/
```
Checkout branch
```bash
git checkout release/v0.1.0
```
Build and install LSP
```bash
./scripts/install.sh
``` 

Base64 encode admin macaroon and TLS cert
```bash
cat ~/.lnd/tls.cert | base64 --wrap=0
# Mainnet
cat ~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon | base64 --wrap=0
# Testnet
cat ~/.lnd/data/chain/bitcoin/testnet/admin.macaroon | base64 --wrap=0
```
Create lsp data and config directory
```bash
mkdir -p ~/.lsp/backups
```
Edit the `~/.lsp/lsp.conf` file
```bash
ACCOUNTING_CURRENCY=EUR
BACKUP_AWS_REGION=eu-central-1
BACKUP_AWS_ACCESS_KEY_ID=
BACKUP_AWS_SECRET_ACCESS_KEY=
BACKUP_S3_BUCKET=satimoto-lspX-testnet-channel-backup
BACKUP_FILE_PATH=/home/ubuntu/.lsp/backups
CIRCUIT_PERCENT=0.5
DB_USER=satimoto
DB_PASS=
DB_HOST=satimoto.cluster-csvwlfckqqfq.eu-central-1.rds.amazonaws.com
DB_NAME=satimoto
FERP_RPC_ADDRESS=ferp.satimoto.service:50000
FCM_API_KEY=
LND_GRPC_HOST=127.0.0.1:10009
LND_TLS_CERT=
LND_MACAROON=
OCPI_RPC_ADDRESS=ocpi.satimoto.service:50000
SCID_CACHE_SIZE=10
PSBT_BATCH_TIMEOUT=30
PBST_HTLC_RESUME_TIMEOUT=20
BASE_FEE_MSAT=0
FEE_RATE_PPM=10
TIME_LOCK_DELTA=100
METRIC_PORT=9102
REST_PORT=9002
RPC_PORT=50000
SHUTDOWN_TIMEOUT=20
```
Add a systemd service to manage LSP. Edit `/etc/systemd/system/lsp.service`
```bash
[Unit]
Description=LSP daemon
After=lnd.service
Wants=lnd.service

[Service]
User=ubuntu
Group=ubuntu
ExecStart=/home/ubuntu/go/bin/lsp --configfile=/home/ubuntu/.lsp/lsp.conf
Restart=on-failure
RestartSec=30

[Install]
WantedBy=multi-user.target
```
Start lsp and enable running on startup
```bash
sudo systemctl start lsp
sudo systemctl enable lsp
```
Edit `~/.profile` and add the lines
```bash
alias loglsp="journalctl -fu lsp"
alias logall="journalctl -fu bitcoind -u lnd -u lsp"
```
