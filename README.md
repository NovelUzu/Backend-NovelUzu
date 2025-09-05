# NovelUzu's backend

## Core dependencies

go version 1.24
gin-gonic (https://github.com/gin-gonic/gin) version 1.10
socket.io (github.com/zishang520/socket.io/v2) version 2.3.8

The remaining dependencies can be found on go.mod

#### Relevant external versions

postgres version 16.9
go server hosted on OpenNebula 6.10

## Deployment

To deploy this project run:

```
go mod tidy
go run main.go

```
Or compile it as a binary:
```
go mod tidy
go build main.go
./main
```

generate documentation:
```
swag init --output config/swagger
```
## Usage/Examples

#### Development server swagger docs

~~~ copy

http://backnoveluzu.eslus.org:8080
~~~

#### Production server swagger docs

~~~ copy

https://localhost:8080
~~~
# Backend-NovelUzu
