package cache

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
)

type Post struct {
	ID      uint64 `db:"id"`
	Title   string `db:"title"`
	Content string `db:"content"`
}

type PostService struct {
	repo       *PostRepository
	natsCache  *PostNatsCache
	redisCache *PostRedisCache
}

func (s *PostService) FindByTitleNoCache(title string) (Post, error) {
	post, err := s.repo.FindByTitle(title)
	if err != nil {
		return Post{}, err
	}
	return post, nil
}

func (s *PostService) FindByTitleCacheByNats(title string) (Post, error) {
	post, err := s.natsCache.FindByTitle(title)
	if err != nil && !errors.Is(err, jetstream.ErrKeyNotFound) {
		return Post{}, err
	} else if err == nil {
		return post, nil
	}

	post, err = s.repo.FindByTitle(title)
	if err != nil {
		return Post{}, err
	}

	err = s.natsCache.Put(post)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func (s *PostService) FindByTitleCacheByRedis(title string) (Post, error) {
	post, err := s.redisCache.FindByTitle(title)
	if err != nil && !errors.Is(err, redis.Nil) {
		return Post{}, err
	} else if err == nil {
		return post, nil
	}

	post, err = s.repo.FindByTitle(title)
	if err != nil {
		return Post{}, err
	}

	err = s.redisCache.Put(post)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

type PostRepository struct {
	db *sqlx.DB
}

func (r *PostRepository) FindByTitle(title string) (Post, error) {
	var p Post
	err := r.db.Get(&p, "SELECT * FROM `post` WHERE `title` = ?", title)
	if err != nil {
		return Post{}, err
	}
	return p, nil
}

type PostNatsCache struct {
	kv jetstream.KeyValue
}

func (c *PostNatsCache) FindByTitle(title string) (Post, error) {
	entry, err := c.kv.Get(context.TODO(), title)
	if err != nil {
		return Post{}, err
	}

	var p Post
	if err := json.Unmarshal(entry.Value(), &p); err != nil {
		return Post{}, err
	}

	return p, nil
}

func (c *PostNatsCache) Put(p Post) error {
	data, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	_, err = c.kv.Put(context.TODO(), p.Title, data)
	if err != nil {
		return err
	}

	return nil
}

type PostRedisCache struct {
	rdb *redis.Client
}

func (c *PostRedisCache) FindByTitle(title string) (Post, error) {
	var p Post
	err := c.rdb.HGetAll(context.TODO(), title).Scan(&p)
	if err != nil {
		return Post{}, err
	}
	return p, nil
}

func (c *PostRedisCache) Put(p Post) error {
	return c.rdb.HSet(context.TODO(), p.Title, "id", p.ID, "title", p.Title, "content", p.Content).Err()
}
