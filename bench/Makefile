build:
	go build -o bench .

replace-bench:
	GOOS=linux go build -o bench .
	mv bench ../provisioning/ansible-replace-bench/
