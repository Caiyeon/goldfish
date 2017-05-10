FROM golang:1.8.0

# fetch goldfish
RUN go get github.com/caiyeon/goldfish

# build and run goldfish
WORKDIR $GOPATH/src/github.com/caiyeon/goldfish
RUN go build server.go

# build public files to be served by goldfish
WORKDIR $GOPATH/src/github.com/caiyeon/goldfish/frontend
RUN curl -sL https://deb.nodesource.com/setup_7.x | bash -
RUN apt-get install -y nodejs
RUN npm install -g cross-env
RUN npm install
RUN npm run build

EXPOSE 8000

ENV VAULT_ADDR http://127.0.0.1:8200
ENV VAULT_TOKEN goldfish

WORKDIR $GOPATH/src/github.com/caiyeon/goldfish
RUN mv server /usr/local/goldfish

CMD "/usr/local/goldfish" "-addr=${VAULT_ADDR}" "-approle_path=auth/approle/login" "-config_path=data/goldfish" "-role_id=goldfish" "-token=${VAULT_TOKEN}"