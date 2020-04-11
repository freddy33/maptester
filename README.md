# Test code and environment for benchmarking concurrent maps

# Purpose
While working on QSM and big in memory data set, found out that simple golang hashmap and sync.Map are not suitable for my use case.
Here is a benchmark environment to find the best Map to support insertion of concurrently large amount of data with different level of conflicting keys.

# Needed to dev
It's using [Protobuf](https://developers.google.com/protocol-buffers/) and the [golang serializer](https://github.com/golang/protobuf) .
So I installed it using `brew install protobuf`
Then for go I run `go get -u github.com/golang/protobuf/protoc-gen-go`

# How to run
Helper shell: $ `./run.sh`
Show the amount of file and data: `./run.sh show`
Generate all the data file: `./run.sh gen`
Run all the tests: `./run.sh test`

# Latests full run
Output:
`Did test 2622 out of 2622 reached 100 %`
`All tests - 1915655658: Took 10h11m42.43010365s and 407230 MB alloc`
         
