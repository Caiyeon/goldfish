#!/bin/sh
#Entrypoint script that sets up Vault config required for Goldfish
#API calls were converted from:
#- https://github.com/Caiyeon/goldfish/blob/master/vagrant/policies/goldfish.hcl
#- https://github.com/Caiyeon/goldfish/wiki/Production-Deployment#1-prepare-vault-only-needs-to-be-done-once

# Vault container config
export VAULT_ADDR="http://vault:8200"
export VAULT_TOKEN="goldfish"

#One place for curl options
CURL_OPT="-s -H X-Vault-Token:${VAULT_TOKEN}"

# transit backend and approle auth backend need to be enabled
curl ${CURL_OPT} ${VAULT_ADDR}/v1/sys/mounts/transit -d '{"type":"transit"}'
curl ${CURL_OPT} ${VAULT_ADDR}/v1/sys/auth/approle -d '{"type":"approle"}'

# see the policy file for details
curl ${CURL_OPT} -X PUT ${VAULT_ADDR}/v1/sys/policy/goldfish -d '{"rules": "path \"transit/encrypt/goldfish\" {capabilities = [\"read\",\"update\"]}, path \"transit/decrypt/goldfish\" {capabilities = [\"read\",\"update\"]}, path \"secret/goldfish*\" {capabilities = [\"read\",\"update\"]}"}'
curl ${CURL_OPT} ${VAULT_ADDR}/v1/auth/approle/role/goldfish -d '{"policies":"default,goldfish", "secret_id_num_uses":"1", "secret_id_ttl":"5", "period":"24h"}'
curl ${CURL_OPT} ${VAULT_ADDR}/v1/auth/approle/role/goldfish/role-id -d '{"role_id":"goldfish"}'

# initialize transit key. This is not strictly required but is proper procedure
curl ${CURL_OPT} -X POST ${VAULT_ADDR}/v1/transit/keys/goldfish

# production goldfish needs a generic secret endpoint to hot reload settings from. See Configuration page for details
curl ${CURL_OPT} ${VAULT_ADDR}/v1/secret/goldfish -d '{"DefaultSecretPath":"secret/", "TransitBackend":"transit", "UserTransitKey":"usertransit", "ServerTransitKey":"goldfish", "BulletinPath":"secret/bulletins/"}'

#Generate token to start Goldfish with
WRAPPED_TOKEN=`curl ${CURL_OPT} --header "X-Vault-Wrap-TTL: 20" -X POST ${VAULT_ADDR}/v1/auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token`

/app/goldfish -config=/app/docker.hcl -token=${WRAPPED_TOKEN}
