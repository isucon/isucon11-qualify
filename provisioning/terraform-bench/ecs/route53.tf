data "aws_route53_zone" "xi_isucon_dev" {
  name         = "xi.isucon.dev."
}

resource "aws_route53_record" "jia-auth" {
  zone_id = data.aws_route53_zone.xi_isucon_dev.zone_id
  name    = "jia-auth.${data.aws_route53_zone.xi_isucon_dev.name}"
  type    = "CNAME"
  ttl     = "300"
  records = [aws_lb.isucon11q-ecs.dns_name]
  depends_on = [aws_lb.isucon11q-ecs]
}
