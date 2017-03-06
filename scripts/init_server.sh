TOKEN="$(vault write -f -wrap-ttl=20m -format=json \
auth/approle/role/goldfish/secret-id | jq -r .wrap_info.token)"

../server -token="${TOKEN}"