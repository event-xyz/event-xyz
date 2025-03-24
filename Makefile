build-container-image:
	docker build -t anirudhsudhir/eventloop .
run-container:
	sqlite3 event.db "VACUUM;"
	docker run --rm -v ${PWD}/.env:/app/.env -v ${PWD}/event.db:/app/event.db -p 8080:8080 eventloop-image
