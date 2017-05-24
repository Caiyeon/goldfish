# this policy enables a user to view and perform actions
# in the "Administration" -> "Mounts" page

# allows user to list mounts
path "sys/mounts" {
	capabilities = ["read"]
}

# allows user to view all mounts' configs
path "sys/mounts/*" {
	capabilities = ["read"]
}
