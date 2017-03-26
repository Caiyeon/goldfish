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

ARG VAULT_ADDR
ARG VAULT_TOKEN
ENV VAULT_ADDR ${VAULT_ADDR}
ENV VAULT_TOKEN ${VAULT_TOKEN}

CMD ["./server", "-addr=${VAULT_ADDR}", "-token=${VAULT_TOKEN}"]