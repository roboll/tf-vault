resource aws_security_group vault {
    name = "${var.env}-vault"
    description = "vault instances"

    vpc_id = "${var.vpc}"
}

resource aws_security_group_rule vault_ingress_tcp {
    security_group_id = "${aws_security_group.vault.id}"
    cidr_blocks = [ "0.0.0.0/0" ]

    type = "ingress"
    from_port = 8200
    to_port = 8200
    protocol = "tcp"
}

resource aws_security_group_rule vault_ingress_tcp_metrics {
    security_group_id = "${aws_security_group.vault.id}"
    cidr_blocks = [ "0.0.0.0/0" ]

    type = "ingress"
    from_port = 9102
    to_port = 9102
    protocol = "tcp"
}

resource aws_security_group_rule vault_ingress_etcd_self {
    security_group_id = "${aws_security_group.vault.id}"
    self = true

    type = "ingress"
    from_port = 2380
    to_port = 2380
    protocol = "tcp"
}

resource aws_security_group_rule vault_egress {
    security_group_id = "${aws_security_group.vault.id}"
    cidr_blocks = [ "0.0.0.0/0" ]

    type = "egress"
    from_port = 0
    to_port = 0
    protocol = "-1"
}

resource aws_security_group vault_elb {
    name = "${var.env}-vault-elb"
    description = "vault elb"

    vpc_id = "${var.vpc}"
}

resource aws_security_group_rule vault_elb_ingress_tcp {
    security_group_id = "${aws_security_group.vault_elb.id}"
    cidr_blocks = [ "0.0.0.0/0" ]

    type = "ingress"
    from_port = 443
    to_port = 443
    protocol = "tcp"
}

resource aws_security_group_rule vault_elb_egress_instance {
    security_group_id = "${aws_security_group.vault_elb.id}"
    source_security_group_id = "${aws_security_group.vault.id}"

    type = "egress"
    from_port = 8200
    to_port = 8200
    protocol = "tcp"
}

resource null_resource network {
    depends_on = [
        "aws_security_group_rule.vault_ingress_tcp",
        "aws_security_group_rule.vault_ingress_tcp_metrics",
        "aws_security_group_rule.vault_ingress_etcd_self",
        "aws_security_group_rule.vault_egress",
        "aws_security_group_rule.vault_elb_ingress_tcp",
        "aws_security_group_rule.vault_elb_egress_instance"
    ]
}
