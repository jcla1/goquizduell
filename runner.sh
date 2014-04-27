while true; do
    QD_COOKIE_FILE="cookie.gob" go run examples/main.go -randGames=-1
    sleep 90
done