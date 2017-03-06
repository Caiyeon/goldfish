path "sys/*" {
  policy = "deny"
}

path "secret/*" {
  policy = "write"
}

path "secret/foo" {
  policy = "read"
  capabilities = ["create", "sudo"]
}

path "secret/super-secret" {
  capabilities = ["deny"]
}