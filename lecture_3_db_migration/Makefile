# Lecture 3 ka Makefile
# "make <command>" se chala sakte ho

# PostgreSQL container start karo
postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

# simple_bank database banao
createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

# simple_bank database delete karo
dropdb:
	docker exec -it postgres12 dropdb simple_bank

# Migration UP: tables banao
migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

# Migration DOWN: tables hatao (undo)
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown
