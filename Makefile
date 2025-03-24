build-container-image:
	docker build -t eventloop .
run-container:
	docker run --rm -e ELOOP_DEV=1 -p 8080:8080 -v ${PWD}/tmp:/app/data/ eventloop
