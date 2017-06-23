# Docker

## Included Dockerfiles

### Dockerfile
Quickly starts up a Vault and Goldfish in dev mode on the same machine.

To launch:
```bash
# Build goldfish docker image
docker build -t goldfish $GOPATH/src/github.com/caiyeon/goldfish/docker

# Run (goldfish will create a dev instance of vault for itself)
docker run --name goldfish -p 8000:8000 goldfish

# Open up http://localhost:8000 in a browser
```

### Dockerfile.backend
Builds a container to run the Goldfish backend, designed to run within a docker-compose stack.

- Downloads versioned binary from [Goldfish Github releases](https://github.com/Caiyeon/goldfish/releases)
- Exposes port 8000
- Defaults to run Goldfish with the `-dev` flag

To build:

```bash
docker build -F Dockerfile.backend -t goldfish:backend
```

### Dockerfile.frontend
Builds a container to run the Goldfish frontend, designed to run within a docker-compose stack.

- Uses [official nginx container](https://hub.docker.com/_/nginx/)
- Downloads versioned `public.zip` from [Goldfish Github releases](https://github.com/Caiyeon/goldfish/releases)
- Configures nginx to serve assets from `/usr/share/nginx/html`
- Configures nginx to use upstream backend Goldfish container with hostname `goldfish`

To build
```bash
docker build -F Dockerfile.frontend -t goldfish:backend
```

## Using docker-compose
The `Dockerfile.backend` and `Dockerfile.frontend` containers are meant to be brought up in a [docker-compose](https://github.com/docker/compose) stack locally.

Stack includes:
- [Official Vault container](https://hub.docker.com/_/vault/) setting default `VAULT_DEV_ROOT_TOKEN_ID=goldfish`
- Goldfish backend container built from `Dockerfile.backend` and uses `entrypoint.sh` to connect and configure Vault
- Goldfish frontend container built from `Dockerfile.frontend` to connect

To bring up the stack:

```bash
docker-compose up
```

When the stack is created, access it at http://localhost:8080 in a browser.
