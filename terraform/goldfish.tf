# see the github wiki for how to generate this token
variable "wrapping_token" {
    type = "string"
    description = "A wrapped approle secret_id for goldfish to bootstrap to vault"
}

variable "goldfish_version" {
    type = "string"
    description = "Version of goldfish to deploy"
    default = "v0.8.0"
}

# configure how goldfish will listen to requests
variable "listener" {
    type = "map"
    description = "Configuration for goldfish listener. See github.com/caiyeon/goldfish/config for more"

    # this default will NOT launch. It only serves as a template
    default = {
        address = ":443"
        tls_cert_file = ""
        tls_key_file = ""
        tls_disable = "0"
        tls_autoredirect = "0"
    }
}

# configure how goldfish will connect to vault
variable "vault" {
    type = "map"
    description = "Configuration for goldfish connection to vault. See github.com/caiyeon/goldfish/config for more"

    # this default will NOT launch. It only serves as a template
    default = {
        address = "https://vault.rocks:8200"
        tls_skip_verify = "0"
        runtime_config = "secret/goldfish"
        approle_login = "auth/approle/login"
        approle_id = "goldfish"
    }
}

# templating the config file with variables above
data "template_file" "config" {
    template = "${file("${path.module}/config.hcl.tpl")}"
    vars {
        listener = "${var.listener}"
        vault = "${var.vault}"
    }
}

# templating the systemd service file
data "template_file" "service" {
    template = "${file("${path.module}/goldfish-service.tpl")}"
    vars {
        WRAPPING_TOKEN = "${var.wrapping_token}"
    }
}

# templating the deployment file
data "template_file" "deploy" {
    template = "${file("${path.module}/deploy.sh.tpl")}"
    vars {
        GOLDFISH_VERSION = "${var.goldfish_version}"
    }
}

# allocate an ec2 instance and deploy goldfish!
resource "aws_instance" "goldfish" {
    # a small instance is enough
    instance_type = "t2.small"

    # for clarity sake
    tags {
        "ec2-vault-goldfish"
    }

    # provision config file
    provisioner "file" {
        source = "${data.template_file.config.rendered}"
        destination = "/etc/goldfish-config.hcl"
    }
    # provision systemd service file
    provision "file" {
        source = "${data.template_file.service.rendered}"
        destination = "/etc/systemd/system/goldfish.service"
    }

    # provision deployment script as user_data
    user_data = "${data.template_file.deploy.rendered}"
}
