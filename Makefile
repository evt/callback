run_tester_service:
	go run -race cmd/tester_service/main.go

lint:
	gofumpt -w -s ./..
	golangci-lint run --fix

