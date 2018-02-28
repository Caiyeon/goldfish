# This policy exists solely to deny the public demo token from revoking itself

# Allow tokens to revoke themselves
path "auth/token/revoke-self" {
    capabilities = ["deny"]
}
