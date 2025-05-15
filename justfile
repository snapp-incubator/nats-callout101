default:
    @just --list

# build
build:
    @echo '{{ BOLD + CYAN }}Building Binary!{{ NORMAL }}'
    go build -o koochooloo ./cmd/nats-callout101

# update go packages
update:
    @cd ./cmd/nats-callout101 && go get -u

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml
