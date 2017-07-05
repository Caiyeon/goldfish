listener "tcp" {
    address          = "vault-ui.io"
    tls_cert_file    = ""
    tls_key_file     = ""
    tls_disable      = 0
    tls_autoredirect = 1
}

vault {
    address         = "http://127.0.0.1:8200"
    tls_skip_verify = 0
    runtime_config  = "secret/goldfish"
    approle_login   = "auth/approle/login"
    approle_id      = "goldfish"
}
