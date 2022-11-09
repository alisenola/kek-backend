# REST API with golang, gin, and gorm  
This project is an exemplary rest api server built with Go.

See [API Spec](./api.md) (modified from [RealWorld API Spec](https://github.com/gothinkster/realworld/tree/master/api) to simplify)  

API Server technology stack is  

- Server code: `golang`
- REST Server: `gin`
- Database: `MySQL` with `golang-migrate` to migrate  
- ORM: `gorm v2`  
- Dependency Injection: `fx`  
- Unit Testing: `go test` and `testify`
- Configuration management: `cobra`

---  

> ## Getting started  

### 1. Start with docker compose  

> ####  build a docker image  

```bash
// docker build -f cmd/server/Dockerfile -t kek-server-v2/kek-server .
$ make build-docker
``` 

> #### run api server with mysql (see docker-compose.yaml)  

```bash
// docker-compose up --force-recreate
$ make compose

$  docker ps -a
CONTAINER ID        IMAGE                        COMMAND                  CREATED             STATUS              PORTS                               NAMES
e01564708984        kek-server-v2/kek-server   "kek-server --co…"   40 seconds ago      Up 39 seconds       0.0.0.0:3000->3000/tcp              kek-server
105cb25b6d3a        mysql:8.0.17                 "docker-entrypoint.s…"   40 seconds ago      Up 39 seconds       0.0.0.0:5432->5432/tcp, 54320/tcp   my-postgres
```  

> #### Check apis  

Run intellij's .http files in `tools/http/sample directory`(./tools/http/sample)  