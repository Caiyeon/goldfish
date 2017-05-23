# [mandatory]
# server's transit key (stores logon tokens)
# NO OTHER POLICY should be able to write to this key
path "transit/encrypt/goldfish" {
  capabilities = ["read", "update"]
}
path "transit/decrypt/goldfish" {
  capabilities = ["read", "update"]
}

# [mandatory] [changable]
# store goldfish run-time settings here
# goldfish hot-reloads from this endpoint every minute
path "secret/goldfish*" {
  capabilities = ["update"]
}