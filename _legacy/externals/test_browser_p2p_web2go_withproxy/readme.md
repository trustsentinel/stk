# websockets peer-2-peer with bridge (tcp repeater)

Documentation: [Canal seguro con TCP proxy](https://github.com/vrandkode/stk.s/wiki/Canal-seguro-con-TCP-proxy)

```
(I)nitiator -----> Pr(o)xy *:8081 ------> (R)esponder *:8080
```

* Client (I) client/ws-client.go
* Server (o) agent/ws-agent.go
* Repeater (R) auth/proxy.go

```
go build -o a *.go && ./a
```

### Requirements
* NoiseSocket library ([source code](../vendor/noisesocket)) [[https://github.com/go-noisesocket/noisesocket]]
* Golang (net/http)
* Golang Websockets (gorilla) [[https://github.com/gorilla/websocket/tree/master/examples/echo]]
