# Creating Conatiners
## goldfish container
The goldfish container is a backend container that connects to a Vault server and proxies requests coming from the nginx-frontend container.

```
docker built -t demandbase/goldfish .
docker push demandbase/goldfish
```

## goldfish frontend containers

Build the intermediate artifact generation container

```
cd frontend
docker built -d demandbase/goldfish:artifacts .
```

Generate the artifacts using the intermediate container

```
docker run -it --rm -v `pwd`/public:/public demandbase/goldfish:artifacts run build
```

Create the nginx frontend container using the generated artifacts in `/public`. This container will setup a nginx to server static assets and send requests going to `/api` to the backend goldfish container. Nginx configuration is in the `goldfish.conf` file.

```
docker build -t demandbase/goldfish:nginx .
docker push demandbase/goldfish:nginx
```

# Deploying stack

Everytime the stack is re-deployed, the vault credentials for goldfish must be rotated and populated in the seed for the .env.export

```
redis-cli -h redis-docker.demandbase.com set /goldfish/development/GOLDFISH_TOKEN `vault write -f -wrap-ttl=20m -format=json auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token`
```

Deploy the stack by using the Rancher job: https://jenkins.demandbase.com/view/Rancher/job/rancher-deploy-goldfish/
