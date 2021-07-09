module github.com/isucon/isucon11-qualify/bench

go 1.16

replace github.com/isucon/isucon11-qualify/extra/initial-data => ../extra/initial-data

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/google/uuid v1.2.0
	github.com/isucon/isucandar v0.0.0-20210706075559-501b2c3ed1da
	github.com/isucon/isucon11-qualify/extra/initial-data v0.0.0-00010101000000-000000000000
	github.com/labstack/echo v3.3.10+incompatible
)
