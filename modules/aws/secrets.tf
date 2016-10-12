resource tls_private_key account_key {
    algorithm = "RSA"
}

resource aws_iam_user dns_user {
    name = "vault-${var.env}-dns-terraform-user"
    path = "/terraform/${var.env}/"
}

resource aws_iam_user_policy dns_user_policy {
    name = "dns-access"
    user = "${aws_iam_user.dns_user.name}"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "route53:ListHostedZonesByName",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "route53:GetChange",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "route53:*",
            "Resource": "arn:aws:route53:::hostedzone/*"
        }
    ]
}
EOF
}

resource aws_iam_access_key dns_user_key {
    user = "${aws_iam_user.dns_user.name}"

    depends_on = [ "aws_iam_user_policy.dns_user_policy" ]
}

resource acme_registration registration {
    email_address = ""
    server_url = "${var.acme_url}"
    account_key_pem = "${tls_private_key.account_key.private_key_pem}"
}

resource acme_certificate vault_cert {
    server_url = "${var.acme_url}"
    account_key_pem = "${tls_private_key.account_key.private_key_pem}"
    registration_url = "${acme_registration.registration.id}"

    common_name = "vault.${var.global_domain}"
    subject_alternative_names = [ "vault.${var.domain}" ]

    dns_challenge {
        provider = "route53"
        config {
            AWS_ACCESS_KEY_ID = "${aws_iam_access_key.dns_user_key.id}"
            AWS_SECRET_ACCESS_KEY = "${aws_iam_access_key.dns_user_key.secret}"
            AWS_DEFAULT_REGION = "${var.region}"
        }
    }
}

resource aws_s3_bucket_object cert {
    bucket = "${aws_s3_bucket.data_backend.bucket}"
    key = "tls/cert.pem"
    content = "${acme_certificate.vault_cert.certificate_pem}${acme_certificate.vault_cert.issuer_pem}"
    kms_key_id = "${var.kms_key}"
}

resource aws_s3_bucket_object privkey {
    bucket = "${aws_s3_bucket.data_backend.bucket}"
    key = "tls/privkey.pem"
    content = "${acme_certificate.vault_cert.private_key_pem}"
    kms_key_id = "${var.kms_key}"
}

resource null_resource uploads {
    depends_on = [
        "aws_s3_bucket_object.cert",
        "aws_s3_bucket_object.privkey"
    ]
}
