package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/frugal"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	natscontainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

const (
	SubjectEventUserRegister = "event.user.register"
)

type EventUserRegister struct {
	UserID       string `frugal:"1,required"`
	RegisterTime int64  `frugal:"2,required"`
}

func Encode[T any](data *T) []byte {
	bytes := make([]byte, frugal.EncodedSize(data))
	_ = must(frugal.EncodeObject(bytes, nil, data))
	return bytes
}

func Decode[T any](bytes []byte) *T {
	var data T
	_ = must(frugal.DecodeObject(bytes, &data))
	return &data
}

func main() {
	ctx := context.Background()

	natsContainer := must(natscontainer.RunContainer(ctx,
		testcontainers.WithImage("nats:2.9"),
	))
	should(natsContainer.IsRunning())

	nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
	should(nc.IsConnected())

	_ = must(nc.Subscribe(SubjectEventUserRegister, func(msg *nats.Msg) {
		event := Decode[EventUserRegister](msg.Data)
		fmt.Printf("%#v\n", event)
	}))

	must0(nc.Publish(SubjectEventUserRegister, Encode[EventUserRegister](&EventUserRegister{
		UserID:       uuid.New().String(),
		RegisterTime: time.Now().UnixMilli(),
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
