# Running with docker-compose

Quickly start up a Vault and Goldfish stack using [docker-compose](https://github.com/docker/compose). This is meant as a template for deploying to different orchestration environments for production use.

This is similar to a [production deployment](https://github.com/Caiyeon/goldfish/wiki/Production-Deployment).

To launch:
```bash
docker-compose up
```

Go to http://localhost:8000 in a browser and log in with token `goldfish`

## Dockerfile.compose
Builds a container to run the Goldfish, designed to run within a docker-compose stack.

- Downloads versioned binary from [Goldfish Github releases](https://github.com/Caiyeon/goldfish/releases) (`GOLDFISH_VERSION` variable)
- Exposes port 8000
- Starts Goldfish with the the config file `docker.hcl` and `VAULT_DEV_ROOT_TOKEN_ID=goldfish`

## Using docker-compose
Use [docker-compose](https://github.com/docker/compose) to deploy a stack locally.

Stack details:
- [Official Vault container](https://hub.docker.com/_/vault/) setting default `VAULT_DEV_ROOT_TOKEN_ID=goldfish`
- Goldfish [Goldfish release](https://github.com/Caiyeon/goldfish/releases) set in `Dockefile.compose`
- Runs with `entrypoint.sh` to configure Vault for Goldfish
  - Runs [production deployment](https://github.com/Caiyeon/goldfish/wiki/Production-Deployment) commands and configures [Goldfish Policy](https://github.com/Caiyeon/goldfish/blob/master/vagrant/policies/goldfish.hcl) using the [Vault HTTP API](https://www.vaultproject.io/api/index.html) instead of the `vault` binary.
- Uses `docker.hcl` for Goldfish configuration
