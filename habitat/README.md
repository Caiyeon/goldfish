# Habitat Plan

This [Habitat](https://www.habitat.sh/) plan is used to streamline the docker container build process.

## Prerequisites

* [Prepare Vault](https://github.com/Caiyeon/goldfish/wiki/Production-Deployment#1-prepare-vault-only-needs-to-be-done-once)

## How to Build

1) Change into this directory
2) Enter the studio `hab studio enter`
3) Execute: `build`
4) Export: `hab pkg export docker caiyeon/goldfish`

You will have a container image you can publish to the [Docker Store](https://store.docker.com/).

## How to Run

1) Place your TLS certificate and private key in the directory where you will run the container
2) Create a script to populate configuration (and generate a new secret-id wrapper token)

Example (generate.sh):
```
#!/bin/bash

cat <<EOF
addr="http://192.168.56.100:8200"
token="$(vault write -f -wrap-ttl=20m -format=json auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token)" 
goldfish_addr=":443"
cert_file="domain.crt"
key_file="domain.key"
dev_mode=false

EOF
```

3) Run the container (I port forwarded to 8443 so I don't have to run as sudo):
```
docker run -it --rm -e HAB_GOLDFISH="$(./generate.sh)" \
    -v $PWD/domain.crt:/hab/svc/goldfish/data/domain.crt \
    -v $PWD/domain.key:/hab/svc/goldfish/data/domain.key \
    -p 8443:443 caiyeon/goldfish
```

4) Celebrate!
```
REPOSITORY                                  TAG                     IMAGE ID            CREATED             SIZE
caiyeon/goldfish                        0.3.0-20170610050243    b650737fc401        19 minutes ago      141 MB
caiyeon/goldfish                        latest                  b650737fc401        19 minutes ago      141 MB
```
