# `day-ahead-prices-notificator`

## Building and running

```shell
docker build -t day-ahead-prices-notificator .
docker run -it --rm -p 8080:8080 --name day-ahead-prices-notificator --env-file .env day-ahead-prices-notificator
```
