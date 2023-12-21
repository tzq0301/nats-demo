package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	natscontainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

func main() {
	ctx := context.Background()

	natsContainer := must(natscontainer.RunContainer(ctx,
		testcontainers.WithImage("nats:2.9"),
	))
	should(natsContainer.IsRunning())

	nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
	should(nc.IsConnected())

	// ------------------------------------------------------------------------

	total := 5
	var wg sync.WaitGroup
	wg.Add(total)

	must(nc.Subscribe("pubsub", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
		wg.Done()
	}))

	for i := 0; i < total; i++ {
		must0(nc.Publish("pubsub", []byte("Hello "+strconv.Itoa(i))))
	}

	wg.Wait()
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
