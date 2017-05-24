# this policy enables a user to view and perform actions
# in the "Secrets" and "Bulletins" pages

# allows user to list default secrets page
path "secret" {
  capabilities = ["list"]
}

# allows user to browse all secrets, including bulletins
path "secret/*" {
  capabilities = ["read", "list"]
}
