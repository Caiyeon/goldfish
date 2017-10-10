# You need go (v1.9), npm (v5), nodejs (v6)
# You will also need cross-env (sudo npm install -g cross-env)

# install & update npm packages
cd frontend
npm install

# compile frontend
npm run build || exit 1
cd ..

# code-ify static assets
go get github.com/GeertJohan/go.rice || exit 1
go get github.com/GeertJohan/go.rice/rice || exit 1
rm -f rice-box.go || exit 1
rice embed-go || exit 1

# compile goldfish binary
env GOOS=linux GOARCH=amd64 go build -o goldfish-linux-amd64 -v github.com/caiyeon/goldfish || exit 1
env GOOS=windows GOARCH=amd64 go build -o goldfish-windows-amd64.exe -v github.com/caiyeon/goldfish || exit 1

# report build
echo 'Successfully built ' $(git describe --always --tags)
