# You need go (v1.8), npm (v3), nodejs (v7)

# generate frontend static assets
cd frontend
npm run build
cd ..

# code-ify static assets
go get github.com/GeertJohan/go.rice
go get github.com/GeertJohan/go.rice/rice
rice embed-go

# compile goldfish binary
env GOOS=linux GOARCH=amd64 go build -o goldfish-linux-amd64 -v github.com/caiyeon/goldfish
