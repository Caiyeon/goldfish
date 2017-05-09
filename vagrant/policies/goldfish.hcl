# [mandatory]
# server's transit key (stores logon tokens)
# NO OTHER POLICY should be able to read this
path "transit/encrypt/goldfish" {
  policy = "sudo"
}
path "transit/decrypt/goldfish" {
  policy = "sudo"
}

# [mandatory] [changable]
# store goldfish run-time settings here
path "data/goldfish*" {
  policy = "read"
}