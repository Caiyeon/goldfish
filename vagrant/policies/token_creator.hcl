# this policy enables a user to view and perform actions
# in the "Tools" -> "Token Creator" page

# listing and reading roles are not required
# but the make for a better experience for the demo
path "auth/token/roles/" {
	capabilities = ["list"]
}
path "auth/token/roles/*" {
	capabilities = ["read"]
}

# # allows user to create tokens
# # disabled on demo
# path "auth/token/create" {
# 	capabilities = ["update"]
# }
# path "auth/token/create/*" {
# 	capabilities = ["update"]
# }
