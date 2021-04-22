run:
	docker-compose up  --remove-orphans --build

run_tester:
	HTTP_ADDR=:9010 go run -race cmd/tester/main.go

run_callback:
	HTTP_ADDR=:9090 go run -race cmd/callback/main.go

lint:
	gofumpt -w -s ./..
	golangci-lint run --fix

