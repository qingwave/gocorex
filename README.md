# Gocorex

Gocorex is a collection golang useful utils for distributed system and microservices.

## Features

### Redis
- [Distributed Bloom](x/bloom), implemented by redis 
- [Distributed RateLimiter](x/rate), implemented by redis
- [Distributed Lock](x/mutex), implemented by redis

### Cron
- [Cron with min-heap](x/cron/cron.go). implemented by minimal heap
- [TimeWheel](x/cron/timewheel.go)

### Data structures
- [Set](x/containerx/set.go), hash set
- [Heap](x/containerx/heap.go)

### utils
- [trace](x/trace), recoding the latency of operations
- [retry](x/retry), retry operation on conditional
