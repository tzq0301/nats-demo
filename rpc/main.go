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

	nServer := 3
	for i := 0; i < nServer; i++ { // 模拟多实例部署
		nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
		should(nc.IsConnected())

		serverID := i
		must(nc.QueueSubscribe("rpc", "demo-server", func(msg *nats.Msg) {
			fmt.Printf("LoadBalance Trigger: server-%d\n", serverID)
			num := must(strconv.Atoi(string(msg.Data)))
			must0(msg.Respond([]byte(strconv.Itoa(num + 1))))
		}))
	}

	nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
	should(nc.IsConnected())

	nRequest := 15
	var wg sync.WaitGroup
	wg.Add(nRequest)
	for i := 0; i < nRequest; i++ {
		go func(num int) {
			resp := must(nc.RequestWithContext(ctx, "rpc", []byte(strconv.Itoa(num))))
			retNum := must(strconv.Atoi(string(resp.Data)))
			should(num+1 == retNum)
			wg.Done()
		}(i)
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
