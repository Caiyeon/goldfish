### Experimental Docker Image

This docker build is currently experimental.

https://hub.docker.com/r/caiyeon/goldfish/

```bash
# 1. Build the docker image
docker build -t caiyeon/goldfish:<version> .
# or: pull the image from docker hub
docker pull caiyeon/goldfish:<version>

# 2. Create a config.hcl with your desired configuration (see wiki)
# set the file to an environment variable to be passed to docker later
export GOLDFISH_CONFIG=$(cat config.hcl)

# 3. Generate a token (or skip if you want to bootstrap goldfish later)
export VAULT_TOKEN=<see wiki for details>

# 4. Set a port to expose
export GOLDFISH_PORT=8000

# 5. Run (note double quotation marks around config env - preserves newlines)
docker run -it --rm -p ${GOLDFISH_PORT}:${GOLDFISH_PORT} \
    -e GOLDFISH_PORT=${GOLDFISH_PORT} \
    -e GOLDFISH_CONFIG="${GOLDFISH_CONFIG}" \
    -e VAULT_TOKEN=${VAULT_TOKEN} \
    caiyeon/goldfish:<version>
```

---

To run in standalone dev mode:

Note: this will NOT work in OSX due to network being inside the docker VM
```bash
docker pull caiyeon/goldfish:<version>
docker run -it --rm --network=host caiyeon/goldfish:<version>
```
