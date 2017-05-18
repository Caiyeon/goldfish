# Warning

This docker image is under construction. Below are manual steps to run a vault and a goldfish container side by side in the same network.

If you have docker experience, a pull request with a docker stack or docker compose would be greatly appreciated.

```bash
# Make a network
docker network create goldfishnetwork

# Run a dev instance of vault
docker run --name vault --net goldfishnetwork -p 127.0.0.1:8200:8200 --cap-add=IPC_LOCK \
-e 'VAULT_DEV_ROOT_TOKEN_ID=goldfish' vault:0.7.0 &

# Supply vault with necessary data
# Assuming vault client is installed on host machine
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=goldfish
vault policy-write goldfish $GOPATH/src/github.com/caiyeon/goldfish/vagrant/policies/goldfish.hcl

vault mount transit
vault write -f transit/keys/goldfish

vault auth-enable approle
vault write auth/approle/role/goldfish role_name=goldfish secret_id_ttl=5m \
token_ttl=480h token_max_ttl=720h secret_id_num_uses=1 policies=default,goldfish
vault write auth/approle/role/goldfish/role-id role_id=goldfish

# On another terminal, build goldfish docker image
docker build -t goldfish $GOPATH/src/github.com/caiyeon/goldfish/docker

# Generate a wrapped token from approle for goldfish to start with
WRAPPED_TOKEN=$(vault write -f -wrap-ttl=20m -format=json \
auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token)

# Start goldfish image in a container in the same network
docker run --name goldfish --net goldfishnetwork -p 8000:8000 \
-e "VAULT_ADDR=http://vault:8200" \
-e "VAULT_TOKEN=$WRAPPED_TOKEN" \
goldfish
```