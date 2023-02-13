.PHONY: all server migration docker load-images

docker: 
	docker buildx build --platform linux/amd64/v2 -t crossword-server server
	docker buildx build --platform linux/amd64/v2 -t crossword-frontend-server frontend-next
	docker save -o crossword-server.tar crossword-server
	docker save -o crossword-frontend-server.tar crossword-frontend-server
	scp crossword-server.tar theodoreliu@34.135.192.137:crossword-server.tar
	scp crossword-frontend-server.tar theodoreliu@34.135.192.137:crossword-frontend-server.tar

load-images:
	docker load -i ../crossword-server.tar
	docker load -i ../crossword-frontend-server.tar

server:
	cd server && go build && ./server

migration:
	migrate create -ext sql -dir migrations -seq migration
