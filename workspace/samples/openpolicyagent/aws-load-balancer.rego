package system

patch[content] {
	is_create_or_update

	is_kind("Service")
	is_name("greenplum")
	is_namespace("default")
	has_label("app", "greenplum")

	content := add_annotation("service.beta.kubernetes.io/aws-load-balancer-internal", "true")
}

patch[content] {
	is_create_or_update

	is_kind("Service")
	is_name("greenplum")
	is_namespace("default")
	has_label("app", "greenplum")

	content := replace_at_jsonpath("/spec", "loadBalancerSourceRanges", ["10.32.4.1/32"])
}
