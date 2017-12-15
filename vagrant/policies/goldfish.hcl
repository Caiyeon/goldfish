# [mandatory]
# store goldfish run-time settings here
# goldfish hot-reloads from this endpoint every minute
path "secret/goldfish" {
  capabilities = ["read", "update"]
}


# [optional]
# to enable transit encryption, see wiki for details
path "transit/encrypt/goldfish" {
  capabilities = ["read", "update"]
}
path "transit/decrypt/goldfish" {
  capabilities = ["read", "update"]
}


# [optional]
# for goldfish to fetch certificates from PKI backend
path "pki/issue/goldfish" {
  capabilities = ["update"]
}
