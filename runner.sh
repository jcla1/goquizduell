while true; do
    QD_COOKIE_FILE="cookie.gob" go run examples/main.go -constGames=100
    sleep 60
done