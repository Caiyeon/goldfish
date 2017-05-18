echo 'Populating vault with required data'
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$VAULT_ROOT_TOKEN

vault auth $VAULT_ROOT_TOKEN && echo 'Successfully authenticated' \
|| (echo 'Failed to authenticate with vault' && exit 1)


# 1) goldfish requires this policy and the default policy to function properly
vault policy-write goldfish /vagrant/policies/goldfish.hcl


# 2) goldfish requires at least one transit key (to encrypt user credentials)
vault mount transit
vault write -f transit/keys/goldfish


# 3) for security reasons, goldfish strictly requires approle to run (no raw tokens)
vault auth-enable approle

vault write auth/approle/role/goldfish role_name=goldfish secret_id_ttl=5m token_ttl=480h \
token_max_ttl=720h secret_id_num_uses=1 policies=default,goldfish

vault write auth/approle/role/goldfish/role-id role_id=goldfish


# 4) must write these settings unless goldfish launches with "-dev"
vault write secret/goldfish DefaultSecretPath="secret/" TransitBackend="transit" \
UserTransitKey="usertransit" ServerTransitKey="goldfish" BulletinPath="secret/bulletins/"
