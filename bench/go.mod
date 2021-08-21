module github.com/isucon/isucon11-qualify/bench

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/francoispqt/gojay v1.2.13
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0
	github.com/isucon/isucandar v0.0.0-20210821075549-ee64d0785035
	github.com/isucon/isucon10-portal v0.0.0-20201008112716-8c0b637e1bd8
	github.com/isucon/isucon11-qualify/bench/random v0.0.0-00010101000000-000000000000
	github.com/labstack/echo/v4 v4.5.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/pierrec/xxHash v0.1.5
	github.com/pkg/profile v1.6.0
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	golang.org/x/crypto v0.0.0-20210813211128-0a44fdfbc16e // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sys v0.0.0-20210816071009-649d0fc2fce7 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/isucon/isucon11-qualify/bench/random => ./random
