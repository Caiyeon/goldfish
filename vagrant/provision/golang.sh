#!/bin/bash

echo 'Downloading go...'
curl -s https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz -o go.tar.gz
sudo tar -xf go.tar.gz
rm go.tar.gz
sudo rm -r /usr/local/go
sudo mv go /usr/local/

echo 'Setting vagrant user env vars...'
cat /home/vagrant/.profile | grep "GOROOT" || \
echo "export GOROOT=/usr/local/go" >> /home/vagrant/.profile

mkdir /home/vagrant/go
cat /home/vagrant/.profile | grep "GOPATH" || \
echo 'export GOPATH=/home/vagrant/go
export PATH=$PATH:/usr/local/go/bin:/vagrant/bin' >> /home/vagrant/.profile

export GOROOT=/usr/local/go
export GOPATH=/home/vagrant/go
export PATH=$PATH:/usr/local/go/bin:/vagrant/bin

echo 'Downloading goldfish...'
go get github.com/caiyeon/goldfish
cd $GOPATH/src/github.com/caiyeon/goldfish

sudo chown -R vagrant $GOPATH
