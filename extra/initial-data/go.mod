module github.com/isucon/isucon11-qualify/extra/initial-data

go 1.16

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.2.0
	github.com/isucon/isucon11-qualify/bench v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.3.4
)

replace github.com/isucon/isucon11-qualify/bench => ../../bench
