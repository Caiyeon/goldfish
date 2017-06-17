#!/bin/bash

# Goldfish web server (backend)
echo 'Setting up goldfish server as system service...'
echo "[Unit]
Description=Goldfish
After=vault.service

[Service]
Environment='GOMAXPROCS=`nproc`' 'VAULT_TOKEN=$VAULT_ROOT_TOKEN' 'GOPATH=/home/vagrant/go'
WorkingDirectory=/home/vagrant/go/src/github.com/caiyeon/goldfish

Restart=on-failure
User=root
Group=root

ExecStartPre=/usr/local/go/bin/go install github.com/caiyeon/goldfish
ExecStart=/bin/bash -c '/home/vagrant/go/bin/goldfish -config=/vagrant/config.hcl -vault_token=\$(\
/usr/bin/vault write -address=http://127.0.0.1:8200 -f -wrap-ttl=20m -format=json auth/approle/role/goldfish/secret-id | \
/usr/bin/jq -r .wrap_info.token)'

ExecReload=/bin/kill -s=SIGHUP $MAINPID
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/goldfish.service

echo 'Launching goldfish...'
systemctl daemon-reload
systemctl start goldfish



# Required modules
echo 'Installing nodejs and npm...'
sudo apt-get purge -y nodejs npm
curl -sL https://deb.nodesource.com/setup_7.x | sudo -E bash -
sudo apt-get install -y nodejs
echo Nodejs version:
nodejs -v
echo NPM version:
npm -v

echo 'Installing node modules...'
cd /home/vagrant/go/src/github.com/caiyeon/goldfish/frontend
sudo npm install -g cross-env
sudo npm install



# Frontend with hot reload
echo 'Setting for goldfish frontend as system service with hot reload...'
echo "[Unit]
Description=Goldfish Frontend

[Service]
Restart=on-failure
User=root
Group=root
WorkingDirectory=/home/vagrant/go/src/github.com/caiyeon/goldfish/frontend
ExecStartPre=/usr/bin/npm update
ExecStart=/usr/bin/npm run dev
ExecReload=/bin/kill -s=SIGHUP \$MAINPID
KillSignal=SIGTERM
TimeoutSec=300

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/goldfish_frontend.service

echo 'Launching goldfish frontend with hot reload...'
systemctl daemon-reload
systemctl start goldfish_frontend
