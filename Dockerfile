FROM golang:1.8
MAINTAINER devops@demandbase.com

# fetch goldfish
RUN go get github.com/Demandbase/goldfish

# build and run goldfish
WORKDIR $GOPATH/src/github.com/Demandbase/goldfish

RUN go build server.go && \
    mv server /usr/local/bin/goldfish

EXPOSE 8000

ENV VAULT_ADDR=http://127.0.0.1:8200 \
    VAULT_TOKEN=goldfish

CMD /usr/local/bin/goldfish -dev -vault_addr=${VAULT_ADDR} -vault_token=${VAULT_TOKEN}
