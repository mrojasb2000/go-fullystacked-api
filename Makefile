# database name
DB_NAME ?= postgres

# database type
DB_TYPE ?= postgres

# database username
DB_USER ?= db

# database password
DB_PWD ?= postgres

# psql URL
IP=127.0.0.1

# database port
DB_PORT ?= 15432

PSQLURL ?= $(DB_TYPE)://$(DB_USER):$(DB_PWD)@$(IP):$(DB_PORT)/$(DB_NAME)

# sqlc yaml file
SQLC_YAML ?= ./sqlc.yaml

.PHONY : postgresup postgresdown psql createdb teardown_recreate generate

postgresup:
		docker run --name fullystacked -v $(PWD):/usr/share/userapi -e POSTGRES_PASSWORD=$(DB_PWD) -p $(DB_PORT):5432 -d $(DB_NAME)

postgresdown:
		docker stop fullystacked || true && docker rm fullystacked || true

psql:
		docker exec -it fullystacked psql $(PSQLURL)

# task to create database without typing it manually
createdb:
	docker exec -it fullystacked psql $(PSQLURL) -C "\i /usr/share/userapi/db/schema.sql"

teardown_recreate: postgresdown postgresup
	sleep 5
	$(MAKE) createdb

generate:
	@echo "Generating Go models with sqlc"
	sqlc generate -f $(SQLC_YAML)

