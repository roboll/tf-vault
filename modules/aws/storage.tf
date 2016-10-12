resource aws_s3_bucket data_backend {
    bucket = "vault.${var.domain}"
    acl = "private"

    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::vault.${var.domain}",
            "Action": "s3:ListBucket",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        },
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::vault.${var.domain}/*",
            "Action": "s3:DeleteObject",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        },
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::vault.${var.domain}/*",
            "Action": "s3:GetObject",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        },
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::vault.${var.domain}/*",
            "Action": "s3:PutObject",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        }
    ]
}
EOF

    tags {
        Name = "vault.${var.domain}"
        Environment = "${var.env}"
    }
}

resource aws_s3_bucket audit_backend {
    bucket = "audit.vault.${var.domain}"
    acl = "private"

    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::audit.vault.${var.domain}",
            "Action": "s3:ListBucket",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        },
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::audit.vault.${var.domain}/*",
            "Action": "s3:GetObject",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        },
        {
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::audit.vault.${var.domain}/*",
            "Action": "s3:PutObject",
            "Principal": {
                "AWS": "${aws_iam_role.vault.arn}"
            }
        }
    ]
}
EOF

    tags {
        Name = "audit.vault.${var.domain}"
        Environment = "${var.env}"
    }
}
