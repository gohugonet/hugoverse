[![Tests on Linux, MacOS and Windows](https://github.com/bep/lazycache/workflows/Test/badge.svg)](https://github.com/bep/lazycache/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/lazycache)](https://goreportcard.com/report/github.com/bep/lazycache)
[![codecov](https://codecov.io/github/bep/lazycache/branch/main/graph/badge.svg?token=HJCUCT07CH)](https://codecov.io/github/bep/lazycache)
[![GoDoc](https://godoc.org/github.com/bep/lazycache?status.svg)](https://godoc.org/github.com/bep/lazycache)

**Lazycache** is a simple thread safe in-memory LRU cache. Under the hood it leverages the great [simpleru package in golang-lru](https://github.com/hashicorp/golang-lru), with its exellent performance. One big difference between `golang-lru` and this library is the [GetOrCreate](https://pkg.go.dev/github.com/bep/lazycache#Cache.GetOrCreate) method, which provides: 

* Non-blocking cache priming on cache misses. 
* A guarantee that the prime function is only called once for a given key.
* The cache's [RWMutex](https://pkg.go.dev/sync#RWMutex) is not locked during the execution of the prime function, which should make it easier to reason about potential deadlocks.

Other notable features:

* The API is [generic](https://go.dev/doc/tutorial/generics)
* The cache can be [resized](https://pkg.go.dev/github.com/bep/lazycache#Cache.Resize) while running.
* When the number of entries overflows the defined cache size, the least recently used item gets discarded (LRU).

