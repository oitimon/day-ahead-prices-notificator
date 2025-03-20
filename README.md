# `day-ahead-prices-notificator`
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/6883bd62e8e64ec09e61e819dc4181fb)](https://app.codacy.com/gh/oitimon/day-ahead-prices-notificator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)

## Building and running

```shell
docker build -t day-ahead-prices-notificator .
docker run -it --rm -p 8080:8080 -p 9090:9090 --name day-ahead-prices-notificator --env-file .env day-ahead-prices-notificator
```

## Tests

```shell
go test -v ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
