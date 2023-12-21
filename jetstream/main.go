package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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

	js := must(jetstream.New(nc))

	// ------------------------------------------------------------------------

	s := must(js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"ORDERS.*"},
	}))

	for i := 0; i < 100; i++ {
		must(js.Publish(ctx, "ORDERS.new", []byte("hello message "+strconv.Itoa(i))))
		fmt.Printf("Published hello message %d\n", i)
	}

	c := must(s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "CONS",
		AckPolicy: jetstream.AckExplicitPolicy,
	}))

	messageCounter := 0
	msgs := must(c.Fetch(10))
	for msg := range msgs.Messages() {
		fmt.Printf("Received a JetStream message via fetch: %s\n", string(msg.Data()))
		must0(msg.Ack())
		messageCounter++
	}

	fmt.Printf("received %d messages\n", messageCounter)

	if msgs.Error() != nil {
		fmt.Println("Error during Fetch(): ", msgs.Error())
	}

	cons := must(c.Consume(func(msg jetstream.Msg) {
		fmt.Printf("Received a JetStream message via callback: %s\n", string(msg.Data()))
		must0(msg.Ack())
		messageCounter++
	}))
	defer cons.Stop()

	it := must(c.Messages())
	for i := 0; i < 10; i++ {
		msg := must(it.Next())
		fmt.Printf("Received a JetStream message via iterator: %s\n", string(msg.Data()))
		must0(msg.Ack())
		messageCounter++
	}
	it.Stop()

	for messageCounter < 100 {
		time.Sleep(10 * time.Millisecond)
	}

	kv := must(js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: "kv",
	}))
	must(kv.Put(ctx, "foo", []byte("bar")))
	value := string(must(kv.Get(ctx, "foo")).Value())
	println(value)
	should(value == "bar")
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
