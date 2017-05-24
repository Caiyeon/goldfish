# this policy enables a user to view and perform actions
# in the "Tools" -> "Transit" page

path "transit/encrypt/usertransit" {
	capabilities = ["update"]
}
path "transit/decrypt/usertransit" {
	capabilities = ["update"]
}
