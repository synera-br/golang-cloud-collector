build:
	go build -o server main.go

swag:
	swag init --parseDependency -td "{{,}}" --parseInternal --parseDepth 1 -g cmd/main.go  --output docs/swagger

tidy:
	go mod tidy

run-swag:
	make swag
	go run cmd/*

run:
	go run cmd/*

sqlc:
	cd infra/database/sqlc \
	&& sqlc generate

protobuf:
	mkdir -p pkg \
	&& cd proto \
	&& buf mod update \
	&& buf generate \
	&& cd ..

mocks:
	mockery \
	--dir=internal \
	--dir=pkg \
	--output=tests/repomocks \
	--outpkg=repomocks \
	--all

test:
	go test ./... -cover -v

watch:
	clear
	ulimit -n 1000 
	make swag
	make tidy
	clear
	reflex -s -r '\.go$$' make run
	clear
	watch --chgexit -n 1 "ls --all -l --recursive --full-time | sha256sum"  && echo "Detected the modification of a file or directory"
