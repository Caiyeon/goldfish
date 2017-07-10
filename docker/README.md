This docker image will run a dev instance of vault and goldfish on the same machine.

If you have docker experience, a pull request with a docker stack or docker compose would be greatly appreciated.

To launch:
```bash
# Build goldfish docker image
docker build -t goldfish $GOPATH/src/github.com/caiyeon/goldfish/docker

# Run (goldfish will create a dev instance of vault for itself)
docker run --name goldfish -p 8000:8000 goldfish

# Open up http://localhost:8000 on a browser
```
