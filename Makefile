.PHONY: all server migration

server:
	cd server && go build && ./server

migration:
	migrate create -ext sql -dir migrations -seq migration
