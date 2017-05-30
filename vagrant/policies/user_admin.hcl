# this policy enables a user to view and perform actions
# in the "Administration" -> "Users" page

# allows user to list tokens and their details
path "auth/token/accessors/" {
	capabilities = ["list", "sudo"]
}
path "auth/token/lookup-accessor*" {
	capabilities = ["update", "sudo"]
}

# # allows user to delete tokens
# # disabled for the demo environment
# path "auth/token/revoke-accessor" {
# 	capabilities = ["update"]
# }

# allows user to list users and their details
path "auth/userpass/users/" {
	capabilities = ["list"]
}
path "auth/userpass/users/*" {
	capabilities = ["read"]
	# use the following if you wish to allow deletion:
	# capabilities = ["read", "delete"]
}

# allows users to list approle details
path "auth/approle/role/" {
	capabilities = ["list"]
}
path "auth/approle/role/*" {
	capabilities = ["read"]
	# use the following if you wish to allow deletion:
	# capabilities = ["read", "delete"]
}
