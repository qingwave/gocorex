# Gocorex

Gocorex is a collection golang useful utils for distributed system and microservices.

## Features

### Distributed Bloom Filter
- [Redis Bloom](x/bloom)

### Distributed Rate Limiter
- [Redis RateLimiter](x/rate)

### Distributed Lock
- [Redis Lock](x/syncx/redislock)
- [Etcd Lock](x/syncx/etcdlock)
- [ZooKeeper Lock](x/syncx/zklock)

### Service Discovery
- [Etcd discovery](x/discovery/etcdiscovery/)
- [ZooKeeper discovery](x/discovery/zkdiscovery/)

### PubSub
- [Etcd PubSub](x/pubsub)

### Cron
- [Cron with min-heap](x/cron/cron.go), implemented by minimal heap
- [TimeWheel](x/cron/timewheel.go)

### Concurrency
- [Group](x/syncx/group/group.go), wrap the WaitGroup
- [ErrGroup](x/syncx/group/errgroup.go), run groups of goroutines, context cancel when meet error
- [CtrlGroup](x/syncx/group/ctrlgroup.go), run special number goroutines

### Data structures
- [Set](x/containerx/set.go), hash set with generics
- [Heap](x/containerx/heap.go), heap with generics

### utils
- [trace](x/trace), recoding the latency of operations
- [retry](x/retry), retry operation on conditional
