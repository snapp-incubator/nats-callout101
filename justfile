default:
    @just --list

# build
build:
    @echo '{{ BOLD + CYAN }}Building Binary!{{ NORMAL }}'
    go build -o nats-callout101 ./cmd/nats-callout101

# update go packages
update:
    @cd ./cmd/nats-callout101 && go get -u

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml
