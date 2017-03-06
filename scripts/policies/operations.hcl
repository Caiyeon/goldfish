path "secret/env/prod/*" {
  policy = "read"
}

path "secret/shared/*" {
  policy = "read"
}

path "secret/ssl/private-key" {
  policy = "read"
}