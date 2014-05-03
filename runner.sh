while true; do
    QD_COOKIE_FILE="cookie.gob" go run player/main.go -const-games=115 -ans-stddev=0.7
    sleep 60
done
