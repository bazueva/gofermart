include .env

lint:
	golangci-lint run

test:
	go test ./...

.PHONY: jet-generate
jet-generate:
	jet -dsn="$(DATABASE_DSN)" \
		-schema=public \
		-path=./schema.gen \
		-ignore-tables="schema_migrations"