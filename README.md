# demo

## Getting started

安装 go ctl
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
goctl env check --install --verbose --force

# gen proto
protoc net.proto --go_out=.

# gen goctl new http
goctl api new api --style go_zero
# regen
goctl api go --api ./api.api --dir . -style go_zero

```
test 
```bash
curl --request GET 'http://127.0.0.1:8888/from/me' 

curl -X POST -H "Content-Type: application/json" -d '{"username":"hello", "password":"world"}' 'http://127.0.0.1:8888/v1/user/login' 
```