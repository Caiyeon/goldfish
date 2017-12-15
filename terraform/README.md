### Terraform

This is a terraform module template for goldfish

Sample usage:
```ruby
module "goldfish" {
    # remember to use `terraform get` to fetch the module
    source = "github.com/caiyeon/goldfish//terraform"

    # deployment config variables
    wrapping_token = "See wiki for how to generate this"
    goldfish_version = "v0.8.0"
    listener = {
        address = ":443"
        tls_cert_file = ""
        tls_key_file = ""
        tls_disable = "0"
        tls_autoredirect = "0"
    }
    vault = {
        address = "https://vault.rocks:8200"
        tls_skip_verify = "0"
        runtime_config = "secret/goldfish"
        approle_login = "auth/approle/login"
        approle_id = "goldfish"
    }
}

output "goldfish_public_ip" {
    value = "${module.goldfish.public_ip}"
}
```

### Fineprint

This terraform module will NOT work out of the box (for obvious reasons). You (the operator) will need to comb through each variable and possibly change the value.

In particular, goldfish's certificates are not handled in this module. You may want to add steps to fetch those certificates in `user_data.sh`, or provision them.

It is highly recommended to add steps in `user_data.sh` to disable swap and ssh for security reasons, as goldfish may contain sensitive data in memory for brief moments in transit.
