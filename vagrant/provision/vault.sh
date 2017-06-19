#!/bin/bash

echo 'Setting vagrant user env vars...'
cat /home/vagrant/.profile | grep 'VAULT_ADDR' || \
echo 'export VAULT_ADDR=http://127.0.0.1:8200' >> /home/vagrant/.profile
cat /home/vagrant/.profile | grep "VAULT_TOKEN" || \
echo "export VAULT_TOKEN=$VAULT_ROOT_TOKEN" >> /home/vagrant/.profile

echo 'Building vault...'
export GOROOT=/usr/local/go
export GOPATH=/home/vagrant/go
export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
go get github.com/hashicorp/vault
cd /home/vagrant/go/src/github.com/hashicorp/vault
make bootstrap
make dev
cp bin/vault /usr/bin/

echo 'Setting up vault dev server as system service...'
echo "[Unit]
Description=Vault Server

[Service]
Environment='GOMAXPROCS=`nproc`' 'GOROOT=/usr/local/go' 'GOPATH=/home/vagrant/go'
WorkingDirectory=/home/vagrant/go/src/github.com/hashicorp/vault

Restart=on-failure
User=root
Group=root

ExecStart=/home/vagrant/go/src/github.com/hashicorp/vault/bin/vault server -dev -dev-root-token-id=$VAULT_ROOT_TOKEN
ExecReload=/bin/kill -s=SIGHUP $MAINPID
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/vault.service

echo 'Starting vault...'
systemctl daemon-reload
systemctl start vault
sleep 10

echo 'Populating vault with sample data for goldfish'
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$VAULT_ROOT_TOKEN
vault auth $VAULT_ROOT_TOKEN
