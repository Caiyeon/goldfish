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

**IN DEVELOPMENT**
A docker-compose file has been added, this will allow you to run vault and goldfish in seperate containers as a compose stack.
To run, cd to the goldfish/docker dir;
```
docker-compose up docker-compose.yml -d
```
^^ In order for that to work a docker image will have to built for goldfish that omits the vault content and runs goldfish only. Happy to help with that if required. The compose script may need tweaking with regards to the goldfish service in order to pass it any information it needs about the vault container. 
