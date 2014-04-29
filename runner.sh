while true; do
    QD_COOKIE_FILE="cookie.gob" go run examples/main.go -constGames=115 -ansStdDev=0.7
    sleep 60
done