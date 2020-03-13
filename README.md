# Test code and environment for benchmarking concurrent maps

# Purpose
While working on QSM and big in memory data set, found out that simple golang hashmap and sync.Map are not suitable for my use case.
Here is a benchmark environment to find the best Map to support insertion of concurrently large amount of data with different level of conflicting keys.

# Needed to run
It's using [Protobuf](https://developers.google.com/protocol-buffers/) and the [golang serializer](https://github.com/golang/protobuf) .
So I installed it using `brew install protobuf`
Then for go I run `go get -u github.com/golang/protobuf/protoc-gen-go`

