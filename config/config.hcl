listener "tcp" {
	address       = "127.0.0.1:8000"
	tls_cert_file = ""
	tls_key_file  = ""
	tls_disable   = 1
}

vault {
	address         = "http://127.0.0.1:8200"
	tls_skip_verify = 0
	runtime_config  = "secret/goldfish"
	approle_login   = "auth/approle/login"
	approle_id      = "goldfish"
}
