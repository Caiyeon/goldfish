path "secret/test/*" {
  policy = "write"
}

path "secret/*" {
  policy = "read"
}

path "secret/classified/*" {
  capabilities = ["deny"]
}