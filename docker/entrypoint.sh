#!/bin/sh
#Entrypoint script that sets up Vault config required for Goldfish
#API calls were converted from https://github.com/Caiyeon/goldfish/blob/master/vagrant/policies/goldfish.hcl

# Vault container config
export VAULT_ADDR="http://vault:8200"
export VAULT_TOKEN="goldfish"

# transit backend and approle auth backend need to be enabled
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data '{"type":"transit"}' ${VAULT_ADDR}/v1/sys/mounts/transit
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data '{"type":"approle"}' ${VAULT_ADDR}/v1/sys/auth/approle

# see the policy file for details
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data '{"rules": "path \"transit/encrypt/goldfish\" {capabilities = [\"read\",\"update\"]}, path \"transit/decrypt/goldfish\" {capabilities = [\"read\",\"update\"]}, path \"secret/goldfish*\" {capabilities = [\"read\",\"update\"]}"}' ${VAULT_ADDR}/v1/sys/policy/goldfish
curl -H "X-Vault-Token:${VAULT_TOKEN}" --request POST --data '{"policies":"default,goldfish", "secret_id_ttl":"5", "token_ttl":"480h", "token_max_ttl":"720h", "secret_id_num_uses":"1"}' ${VAULT_ADDR}/v1/auth/approle/role/goldfish
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data '{"role_id":"goldfish"}' ${VAULT_ADDR}/v1/auth/approle/role/goldfish/role-id

# initialize transit key. This is not strictly required but is proper procedure
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST ${VAULT_ADDR}/v1/transit/keys/goldfish

# production goldfish needs a generic secret endpoint to hot reload settings from. See Configuration page for details
curl -H "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data '{"DefaultSecretPath":"\"secret/\"", "TransitBackend":"\"transit\"", "UserTransitKey":"\"usertransit\"", "ServerTransitKey":"\"goldfish\"", "BulletinPath":"\"secret/bulletins/\""}' ${VAULT_ADDR}/v1/secret/goldfish

#Generate token to start Goldfish with
WRAPPED_TOKEN=`curl -H "X-Vault-Token: ${VAULT_TOKEN}" -H "X-Vault-Wrap-TTL: 20" --request POST ${VAULT_ADDR}/v1/auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token`

/app/goldfish -config=/app/docker.hcl -token=${WRAPPED_TOKEN}
