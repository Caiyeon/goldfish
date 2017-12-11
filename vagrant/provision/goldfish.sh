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

ExecStart=/usr/local/go/bin/go run /home/vagrant/go/src/github.com/caiyeon/goldfish/server.go -dev

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
curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install npm@5 -g
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
ExecStart=/usr/bin/npm run dev
ExecReload=/bin/kill -s=SIGHUP \$MAINPID
KillSignal=SIGTERM
TimeoutSec=300

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/goldfish_frontend.service

echo 'Launching goldfish frontend with hot reload...'
systemctl daemon-reload
systemctl start goldfish_frontend
