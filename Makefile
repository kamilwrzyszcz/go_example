create_postgres:
	docker run --name example_postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:14-alpine
createdb:
	docker exec -it example_postgres createdb --username=root --owner=root example_db
create_testdb:
	docker exec -it example_postgres createdb --username=root --owner=root test_db
run_postgres:
	docker start example_postgres
stop_postgres:
	docker stop example_postgres
dropdb:
	docker exec -it example_postgres dropdb -example_db
drop_testdb:
	docker exec -it example_postgres dropdb -test_db
migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/example_db?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/dropdb:?sslmode=disable" -verbose down
sqlc:
	sqlc generate

.PHONY: create_postgres createdb create_testdb run_postgres stop_postgres dropdb drop_testdb migrateup migratedown sqlc