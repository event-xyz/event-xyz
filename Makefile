DBPATH=${PWD}/tmp
HOST_DB_PATH="./data/events.db"

build-container-image:
	docker build -t eventloop .
run-container:
	docker run --rm -e ELOOP_DEV=1 -p 7070:8080 -v ${PWD}/data:/app/data/ eventloop

clean_db:
	-rm ./data/*.db

test_sql:
	@echo "using ${HOST_DB_PATH}"
	sqlite3 ${HOST_DB_PATH} < ./sql/schema.sql
	sqlite3 ${HOST_DB_PATH} < ./sql/checks.sql
