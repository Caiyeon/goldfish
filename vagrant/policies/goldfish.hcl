# [mandatory]
# store goldfish run-time settings here
# goldfish hot-reloads from this endpoint every minute
path "secret/goldfish*" {
  capabilities = ["read", "update"]
}


# [optional]
# to enable transit encryption:
# set 'ServerTransitKey' in runtime settings
# and initialize the key: 'vault write -f transit/keys/goldfish'
# see wiki for details
path "transit/encrypt/goldfish" {
  capabilities = ["read", "update"]
}
path "transit/decrypt/goldfish" {
  capabilities = ["read", "update"]
}
