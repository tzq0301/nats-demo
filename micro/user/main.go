package main

import (
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	nc := must(nats.Connect("nats://127.0.0.1:4222"))
	should(nc.IsConnected())

	// ------------------------------------------------------------------------

	id := must(uuid.NewUUID()).String()
	println(id)

	srv := must(micro.AddService(nc, micro.Config{
		Name:    "UserService",
		Version: "1.0.0",
	}))
	userSvc := srv.AddGroup("svc.user")
	must0(userSvc.AddEndpoint("who", micro.HandlerFunc(func(request micro.Request) {
		must0(request.Respond([]byte(id)))
	})))

	select {}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func must0(err error) {
	if err != nil {
		panic(err)
	}
}

func should(b bool) {
	if !b {
		panic(b)
	}
}
