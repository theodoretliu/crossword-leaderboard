.PHONY: all server migration docker docker-server docker-frontend load-images

docker-server:
	docker buildx build --platform linux/amd64/v2 -t crossword-server server
	docker save -o crossword-server.tar crossword-server
	scp crossword-server.tar theodoreliu@crossword.theodoretliu.com:

docker-frontend:
	docker buildx build --platform linux/amd64/v2 -t crossword-frontend-server frontend-next
	docker save -o crossword-frontend-server.tar crossword-frontend-server
	scp crossword-frontend-server.tar theodoreliu@crossword.theodoretliu.com:

docker: docker-server docker-frontend

load-images:
	docker load -i ../crossword-server.tar
	docker load -i ../crossword-frontend-server.tar

server:
	cd server && go build && ./server

migration:
	migrate create -ext sql -dir migrations -seq migration
