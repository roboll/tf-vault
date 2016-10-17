variable env {}
variable region {}

variable vpc {}
variable subnets { type = "list" }

variable kms_key {}
variable ssh_keypair {}

variable domain {}
variable global_domain {}
variable security_group {}

variable vpc_dns_zone_id {}
variable public_dns_zone_id {}
variable routing_policy { default = "PRIMARY" }

variable image_id {}
variable replicas { default = 3 }
variable instance_type { default = "t2.small" }
variable ebs_optimized { default = false }

variable root_volume_type { default = "gp2" }
variable root_volume_size { default = 20 }

variable acme_url { default = "https://acme-v01.api.letsencrypt.org/directory" }

variable vault_image { default = "quay.io/roboll/vault:v0.4.0" }
variable vault_metrics_image { default = "quay.io/roboll/vault-metrics:v0.4.0" }
variable vault_ssh_image { default = "quay.io/roboll/vault-ssh-coreos:v0.2.0" }

provider aws {
    region = "${var.region}"
}

provider kms {
    region = "${var.region}"
}

resource aws_iam_role vault {
    name = "${var.env}-vault"
    path = "/${var.env}/"
    assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            }
        }
    ]
}
EOF

    provisioner local-exec { command = "sleep 60" }
}

resource aws_iam_role_policy vault_s3_data {
    name = "${var.env}-vault-s3-data"
    role = "${aws_iam_role.vault.id}"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "${aws_s3_bucket.data_backend.arn}",
            "Action": [ "s3:ListBucket" ]
        },
        {
            "Effect": "Allow",
            "Resource": "${aws_s3_bucket.data_backend.arn}/*",
            "Action": [
                "s3:DeleteObject",
                "s3:GetObject",
                "s3:PutObject"
            ]
        }
    ]
}
EOF
}

resource aws_iam_role_policy vault_s3_audit {
    name = "${var.env}-vault-s3-audit"
    role = "${aws_iam_role.vault.id}"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "${aws_s3_bucket.audit_backend.arn}",
            "Action": [ "s3:ListBucket" ]
        },
        {
            "Effect": "Allow",
            "Resource": "${aws_s3_bucket.audit_backend.arn}/*",
            "Action": [
                "s3:DeleteObject",
                "s3:GetObject",
                "s3:PutObject"
            ]
        }
    ]
}
EOF
}

resource aws_iam_role_policy vault_kms {
    name = "${var.env}-vault-kms"
    role = "${aws_iam_role.vault.id}"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "${var.kms_key}",
            "Action": [
                "kms:Encrypt",
                "kms:Decrypt",
                "kms:ReEncrypt",
                "kms:DescribeKey",
                "kms:GenerateDataKey"
            ]
        },
        {
            "Effect": "Allow",
            "Resource": "*",
            "Action": [ "kms:ListKeys" ]
        }
    ]
}
EOF
}

resource aws_iam_role_policy vault_ec2 {
    name = "${var.env}-vault-ec2"
    role = "${aws_iam_role.vault.id}"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "*",
            "Action": [
                "ec2:DescribeInstances",
                "iam:GetInstanceProfile"
            ]
        }
    ]
}
EOF
}

resource aws_iam_instance_profile vault {
    name = "${var.env}-vault-instance"
    roles = [ "${aws_iam_role.vault.name}" ]

    depends_on = [
        "aws_iam_role_policy.vault_s3_audit",
        "aws_iam_role_policy.vault_s3_data",
        "aws_iam_role_policy.vault_ec2",
        "aws_iam_role_policy.vault_kms",
        "null_resource.network",
        "null_resource.uploads"
    ]

    provisioner local-exec { command = "sleep 60" }
}

resource null_resource instances {
    triggers {
        name = "vault${count.index}"
        hostname = "vault${count.index}.${var.domain}"
    }

    count = "${var.replicas}"
}

resource coreos_cloudconfig cloud_config {
    gzip = true
    template = "${file("${path.module}/config/cloud-config.yaml")}"

    vars {
        etcd_peers = "${join(",",formatlist("%s=http://%s:2380", null_resource.instances.*.triggers.name, null_resource.instances.*.triggers.hostname))}"
        etcd_clients = "${join(",",formatlist("http://%s:2379", null_resource.instances.*.triggers.hostname))}"
        etcd_instance_name = "${element(split(",", join(",", null_resource.instances.*.triggers.name)), count.index)}"

        vault_image = "${var.vault_image}"
        metrics_image = "${var.vault_metrics_image}"
        vault_ssh_image = "${var.vault_ssh_image}"

        fqdn = "vault.${var.domain}"
        region = "${var.region}"

        hostname = "vault${count.index}.${var.domain}"
        data_bucket = "${aws_s3_bucket.data_backend.id}"
        audit_bucket = "${aws_s3_bucket.audit_backend.id}"

        kms_key = "${var.kms_key}"
        ca_cert_pem_b64 = "${base64encode(acme_certificate.vault_cert.issuer_pem)}"
    }

    count = "${var.replicas}"
}

resource aws_instance vault {
    ami = "${var.image_id}"
    instance_type = "${var.instance_type}"

    key_name = "${var.ssh_keypair}"
    user_data = "${element(coreos_cloudconfig.cloud_config.*.rendered, count.index)}"

    iam_instance_profile = "${aws_iam_instance_profile.vault.name}"

    subnet_id = "${element(var.subnets, count.index)}"
    vpc_security_group_ids = [
        "${var.security_group}",
        "${aws_security_group.vault.id}"
    ]

    ebs_optimized = "${var.ebs_optimized}"
    root_block_device {
        volume_type = "${var.root_volume_type}"
        volume_size = "${var.root_volume_size}"
    }

    tags {
        Name = "${var.env}-vault${count.index}"
        Environment = "${var.env}"
    }

    lifecycle { ignore_changes = [ "ami" ] }

    count = "${var.replicas}"
}

resource aws_elb vault {
    name = "${replace(var.env, ".", "-")}-vault"

    internal = true
    subnets = [ "${var.subnets}" ]
    security_groups = [ "${aws_security_group.vault_elb.id}" ]

    instances = [ "${aws_instance.vault.*.id}" ]

    connection_draining = true
    connection_draining_timeout = 60

    listener {
        instance_port = 8200
        instance_protocol = "tcp"
        lb_port = 443
        lb_protocol = "tcp"
    }

    health_check {
        target = "HTTPS:8200/v1/sys/health"

        healthy_threshold = 2
        unhealthy_threshold = 2

        interval = 15
        timeout = 2
    }

    tags {
        Name = "${var.env}-vault"
        Environment = "${var.env}"
    }
}

resource aws_route53_record hostnames {
    zone_id = "${var.vpc_dns_zone_id}"
    name = "${element(split(",", join(",", null_resource.instances.*.triggers.name)), count.index)}.${var.domain}"

    type = "A"
    ttl = "60"

    records = [ "${element(aws_instance.vault.*.private_ip, count.index)}" ]

    count = "${var.replicas}"
}

resource aws_route53_record global {
    zone_id = "${var.public_dns_zone_id}"
    name = "vault.${var.global_domain}"
    type = "A"

    alias {
        name = "${aws_elb.vault.dns_name}"
        zone_id = "${aws_elb.vault.zone_id}"
        evaluate_target_health = true
    }

    set_identifier = "${var.env}"
    failover_routing_policy {
        type = "${var.routing_policy}"
    }
}

output fqdn { value = "${aws_route53_record.global.fqdn}" }
output address { value = "https://${aws_route53_record.global.fqdn}" }

output instance_ips { value = [ "${aws_instance.vault.*.private_ip}" ] }
output instance_hostnames { value = [ "${aws_route53_record.hostnames.*.fqdn}"] }
output instance_addresses { value = [ "${formatlist("https://%s:8200", aws_route53_record.hostnames.*.fqdn)}" ] }

output ca_cert_pem_b64 { value = "${base64encode(acme_certificate.vault_cert.issuer_pem)}" }
