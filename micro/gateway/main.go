package main

import (
	"context"

	"github.com/nats-io/nats.go"
)

func main() {
	ctx := context.Background()

	nc := must(nats.Connect("nats://127.0.0.1:4222"))
	should(nc.IsConnected())

	// ------------------------------------------------------------------------

	resp := must(nc.RequestWithContext(ctx, "svc.user.who", nil))
	println(string(resp.Data))
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
