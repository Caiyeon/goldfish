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

# write some sample policies
vault policy-write admins /vagrant/policies/admins.hcl
vault policy-write developers /vagrant/policies/developers.hcl
vault policy-write operations /vagrant/policies/operations.hcl
vault policy-write readonly /vagrant/policies/readonly.hcl
vault policy-write goldfish /vagrant/policies/goldfish.hcl

# write some sample users
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="admins" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="developers" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="operations" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="" -ttl=720h -renewable=true
vault token-create -policy="jenkins" -no-default-policy=true -ttl=720h -renewable=true -orphan
vault token-create -policy="travis" -no-default-policy=true -ttl=720h -renewable=true -orphan
vault token-create -policy="cron" -no-default-policy=true -ttl=720h -renewable=true -orphan

vault auth-enable userpass
vault write auth/userpass/users/alice password=foo policies="admins" ttl=480h max_ttl=720h
vault write auth/userpass/users/bob password=foo policies="admins" ttl=480h max_ttl=720h
vault write auth/userpass/users/clementine password=foo policies="admins" ttl=480h max_ttl=720h
vault write auth/userpass/users/doug password=foo policies="developers" ttl=480h max_ttl=720h
vault write auth/userpass/users/ethan password=foo policies="developers" ttl=480h max_ttl=720h
vault write auth/userpass/users/fred password=foo policies="developers" ttl=480h max_ttl=720h
vault write auth/userpass/users/gabrielle password=foo policies="operations" ttl=360h max_ttl=720h
vault write auth/userpass/users/hugh password=foo policies="operations" ttl=360h max_ttl=720h
vault write auth/userpass/users/illia password=foo policies="operations" ttl=360h max_ttl=720h
vault write auth/userpass/users/jude password=foo policies="" ttl=360h max_ttl=720h
vault write auth/userpass/users/keith password=foo policies="" ttl=360h max_ttl=720h
vault write auth/userpass/users/larry password=foo policies="" ttl=360h max_ttl=720h
vault write auth/userpass/users/mona password=foo policies="" ttl=360h max_ttl=720h

# transit keys should be initialized
vault mount transit
vault write -f transit/keys/goldfish
vault write -f transit/keys/usertransit

# initialize some sample mount points
vault mount -path=aws -description="Secret backend for amazon web services access credentials generation" aws
vault mount -path=consul -description="Secret backend for consul API token generation" consul
vault mount -path=mssql -description="Secret backend for MS SQL dynamic user credentials generation" mssql
vault mount -path=mysql -description="Secret backend for MySQL dynamic user credentials generation" mysql
vault remount secret/ data/

# write initial goldfish config into vault
vault write data/goldfish DefaultSecretPath="data/" TransitBackend="transit" UserTransitKey="usertransit" \
ServerTransitKey="goldfish" BulletinPath="data/bulletins"

# populate /data/ generic backend with some sample data
vault write data/stardew_valley/crops/spring/blue_jazz buy_price=30 days_to_grow=7 sell_price=50
vault write data/stardew_valley/crops/spring/cauliflower buy_price=80 days_to_grow=12 sell_price=175
vault write data/stardew_valley/crops/spring/coffee_bean buy_price=2500 days_to_grow=10 sell_price=15
vault write data/stardew_valley/crops/spring/garlic buy_price=40 days_to_grow=4 sell_price=60
vault write data/stardew_valley/crops/spring/green_bean buy_price=60 days_to_grow=10 sell_price=40

vault write data/stardew_valley/crops/summer/blueberry buy_price=80 days_to_grow=13 sell_price=50
vault write data/stardew_valley/crops/summer/corn buy_price=150 days_to_grow=14 sell_price=50
vault write data/stardew_valley/crops/summer/hops buy_price=60 days_to_grow=11 sell_price=25
vault write data/stardew_valley/crops/summer/hot_pepper buy_price=40 days_to_grow=5 sell_price=40
vault write data/stardew_valley/crops/summer/melon buy_price=80 days_to_grow=12 sell_price=250

vault write data/stardew_valley/crops/fall/cranberry buy_price=240 days_to_grow=7 sell_price=75
vault write data/stardew_valley/crops/fall/eggplant buy_price=20 days_to_grow=5 sell_price=60
vault write data/stardew_valley/crops/fall/grape buy_price=60 days_to_grow=10 sell_price=80
vault write data/stardew_valley/crops/fall/pumpkin buy_price=100 days_to_grow=13 sell_price=320
vault write data/stardew_valley/crops/fall/yam buy_price=60 days_to_grow=10 sell_price=160

vault write data/stardew_valley/fruit_trees/spring/apricot buy_price=2000 days_to_grow=28 sell_price=50
vault write data/stardew_valley/fruit_trees/spring/cherry buy_price=3400 days_to_grow=28 sell_price=80
vault write data/stardew_valley/fruit_trees/summer/orange buy_price=4000 days_to_grow=28 sell_price=100
vault write data/stardew_valley/fruit_trees/summer/peach buy_price=6000 days_to_grow=28 sell_price=140
vault write data/stardew_valley/fruit_trees/fall/apple buy_price=4000 days_to_grow=28 sell_price=100
vault write data/stardew_valley/fruit_trees/fall/pomegranate buy_price=6000 days_to_grow=28 sell_price=140

vault write data/stardew_valley/animals/coop_animals/chicken buy_price=800 produces="egg, large egg" sell_price=400
vault write data/stardew_valley/animals/coop_animals/duck buy_price=4000 produces="duck egg, duck feather" sell_price=2000
vault write data/stardew_valley/animals/coop_animals/rabbit buy_price=8000 produces="wool, rabbit's foot" sell_price=4000
vault write data/stardew_valley/animals/coop_animals/dinosaur buy_price=-1 produces="dinosaur egg" sell_price=1000

vault write data/stardew_valley/animals/barn_animals/cow buy_price=1500 produces="milk, large milk" sell_price=750
vault write data/stardew_valley/animals/barn_animals/goat buy_price=4000 produces="goat milk, large goat milk" sell_price=2000
vault write data/stardew_valley/animals/barn_animals/sheep buy_price=8000 produces="wool" sell_price=4000
vault write data/stardew_valley/animals/barn_animals/pig buy_price=16000 produces="truffle" sell_price=8000

# configure approle for goldfish server
vault auth-enable approle
vault write auth/approle/role/goldfish role_name=goldfish secret_id_ttl=5m token_ttl=480h \
token_max_ttl=720h secret_id_num_uses=1 policies=default,goldfish
vault write auth/approle/role/goldfish/role-id role_id=goldfish
