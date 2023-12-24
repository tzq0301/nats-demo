package cache

import (
	"context"
	_ "embed"
	"fmt"
	"math/rand"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	mysqlcontainer "github.com/testcontainers/testcontainers-go/modules/mysql"
	natscontainer "github.com/testcontainers/testcontainers-go/modules/nats"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

//go:embed schema.sql
var ddl string

func insertPosts(db *sqlx.DB) {
	db.MustExec(ddl)

	for i := 0; i < 10000; i++ {
		db.MustExec("INSERT INTO `post`(`title`, `content`) VALUES (?, ?)",
			fmt.Sprintf("title-%d", i),
			fmt.Sprintf("content-%d", i))
	}
}

func BenchmarkNoCache(b *testing.B) {
	ctx := context.Background()

	natsContainer := must(natscontainer.RunContainer(ctx, testcontainers.WithImage("nats:2.9")))
	nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
	js := must(jetstream.New(nc))

	mysqlContainer := must(mysqlcontainer.RunContainer(ctx, testcontainers.WithImage("mysql:8")))
	db := must(sqlx.Connect("mysql", must(mysqlContainer.ConnectionString(ctx))))

	insertPosts(db)

	repo := PostRepository{
		db: db,
	}

	cache := PostNatsCache{
		kv: must(js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket: "post",
		})),
	}

	service := PostService{
		repo:      &repo,
		natsCache: &cache,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.FindByTitleNoCache(fmt.Sprintf("title-%d", rand.Int63n(800)))
		if err != nil {
			b.Fatal()
		}
	}
}

func BenchmarkNatsCache(b *testing.B) {
	ctx := context.Background()

	natsContainer := must(natscontainer.RunContainer(ctx, testcontainers.WithImage("nats:2.9")))
	nc := must(nats.Connect(must(natsContainer.ConnectionString(ctx))))
	js := must(jetstream.New(nc))

	mysqlContainer := must(mysqlcontainer.RunContainer(ctx, testcontainers.WithImage("mysql:8")))
	db := must(sqlx.Connect("mysql", must(mysqlContainer.ConnectionString(ctx))))

	insertPosts(db)

	repo := PostRepository{
		db: db,
	}

	cache := PostNatsCache{
		kv: must(js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket: "post",
		})),
	}

	service := PostService{
		repo:      &repo,
		natsCache: &cache,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.FindByTitleCacheByNats(fmt.Sprintf("title-%d", rand.Int63n(800)))
		if err != nil {
			b.Fatal()
		}
	}
}

func BenchmarkRedisCache(b *testing.B) {
	ctx := context.Background()

	mysqlContainer := must(mysqlcontainer.RunContainer(ctx, testcontainers.WithImage("mysql:8")))
	db := must(sqlx.Connect("mysql", must(mysqlContainer.ConnectionString(ctx))))

	redisContainer := must(rediscontainer.RunContainer(ctx, testcontainers.WithImage("redis:6")))
	rdb := redis.NewClient(&redis.Options{
		Addr: must(redisContainer.ConnectionString(ctx))[len("redis://"):],
	})

	insertPosts(db)

	repo := PostRepository{
		db: db,
	}

	redisCache := PostRedisCache{
		rdb: rdb,
	}

	service := PostService{
		repo:       &repo,
		redisCache: &redisCache,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.FindByTitleCacheByRedis(fmt.Sprintf("title-%d", rand.Int63n(800)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
