# Gocorex

Gocorex is a collection golang useful utils for distributed system and microservices.

## Features

### Distributed Bloom Filter
- [Redis Bloom](bloom)

### Distributed Rate Limiter
- [Redis RateLimiter](rate)

### Distributed Lock
- [Redis Lock](syncx/redislock)
- [Etcd Lock](syncx/etcdlock)
- [ZooKeeper Lock](syncx/zklock)

### Service Discovery
- [Etcd discovery](discovery/etcdiscovery/)
- [ZooKeeper discovery](discovery/zkdiscovery/)

### PubSub
- [Etcd PubSub](pubsub)

### Cron
- [Cron with min-heap](cron/cron.go), implemented by minimal heap
- [TimeWheel](cron/timewheel.go)

### Concurrency
- [Group](syncx/group/group.go), wrap the WaitGroup
- [ErrGroup](syncx/group/errgroup.go), run groups of goroutines, context cancel when meet error
- [CtrlGroup](syncx/group/ctrlgroup.go), run special number goroutines

### Data structures
- [Set](containerx/set.go), hash set with generics
- [Heap](containerx/heap.go), heap with generics

### utils
- [trace](trace), recoding the latency of operations
- [retry](retry), retry operation on conditional
