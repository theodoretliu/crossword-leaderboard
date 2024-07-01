.PHONY: all server migration docker docker-server docker-frontend load-images

all: docker-backend docker-frontend

docker-backend:
	docker buildx build --platform linux/amd64/v2 -t crossword-server backend
	docker save -o crossword-server.tar crossword-server
	scp crossword-server.tar theodoreliu@api.crossword.theodoretliu.com:

load-images:
	docker load -i ../crossword-server.tar

server:
	cd server && go build && ./server

migration:
	migrate create -ext sql -dir migrations -seq migration
