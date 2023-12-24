# Benchmark

## NoCache

```bash
go test -bench=BenchmarkNocache
```

```text
     321           3958659 ns/op
PASS
ok      cache   71.788s
```

## Redis as Cache

```bash
go test -bench=BenchmarkRedisCache
```

```text
    3402            399080 ns/op
PASS
ok      cache   61.246s
```

## NATS as Cache

```bash
go test -bench=BenchmarkNatsCache
```

```text
         259           5432904 ns/op
PASS
ok      cache   100.576s
```
