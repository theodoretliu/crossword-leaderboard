#! /usr/local/bin/zsh
(pkill -f "\.\/server" || true) && go build && DB_URL=production.sqlite3 ./server &
