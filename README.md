# Datastore Wrapper [![Go Documentation](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godoc] [![CircleCI](https://circleci.com/gh/mercari/datastore.svg?style=svg)][circleci]

[godoc]: https://godoc.org/go.mercari.io/datastore
[circleci]: https://circleci.com/gh/mercari/datastore

:construction: This package is unstable :construction: [github repo](https://github.com/mercari/datastore)

(AppEngine | Cloud) Datastore wrapper for Go 👉

Simple.
Happy.
Respect standard library.

```
$ go get -u go.mercari.io/datastore
```

## Feature

### DO

* Wrap `google.golang.org/appengine/datastore` and `cloud.google.com/go/datastore`
    * keep key behavior
    * align to `cloud.google.com/go/datastore` first
* Re-implement datastore package
* Re-implement datastore.SaveStruct & LoadStruct
    * Ignore unmapped property
    * Add PropertyTranslator interface
        * Convert types like mytime.Unix to time.Time and reverse it
        * Rename property like CreatedAt to createdAt or created_at and reverse it
* Re-implement PropertyLoadSaver
    * Pass context.Context to Save & Load method
    * It can catch entity from the cache layers as well as from the datastore
* Add retry feature to each RPC
    * e.g. Retry AllocateID when it failed
* Add cache layer uses without Tx
    * About...
        * Local Cache
        * Memcache
        * etc...
    * Easy to ON/OFF switching
* Add RPC hook layer
    * It makes implements [memvache](https://github.com/vvakame/memvache) like cache strategy
* Add some useful methods
    * ~~GetAsIntIDMap, GetAsStringIDMap~~
    * `aedatastore/TransactionContext`

### DON'T

* have utility functions
* support firestore

## Restriction

* `aedatastore` package
    * When using slice of struct, MUST specified `datastore:",flatten"` option.
        * original (ae & cloud) datastore.SaveStruct have different behaviors.
        * see aeprodtest/main.go `/api/test3`

## How To Use

TODO

## Best Practice

TODO

## Project Navigation

* `aedatastore`
    * Datastore Wrapper implementation for AppEngine Datastore
* `aeprodtest`
    * Test code that works in AppEngine production environment
* `boom`
    * [goon](https://github.com/mjibson/goon) likes interface for this package
* `clouddatastore`
    * Datastore Wrapper implementation for Cloud Datastore
* `internal`
    * internal package
* `testbed`
    * Testbed to check the behavior of AE Datastore and Cloud Datastore
* `testsuite`
    * Test suite for this package
        * [AE Datastore entry point](https://github.com/mercari/datastore/blob/master/aedatastore/testsuite_test.go)
        * [Cloud Datastore entry point](https://github.com/mercari/datastore/blob/master/clouddatastore/testsuite_test.go)

## TODO

* Write tests for namespace
* Implement cache layers
* Retry feature
* Write more documents

## Committers

 * Masahiro Wakame ([@vvakame](https://github.com/vvakame))

## Contribution

Please read the CLA below carefully before submitting your contribution.

https://www.mercari.com/cla/

### Setup environment & Run tests

* requirements
    * [gcloud sdk](https://cloud.google.com/sdk/docs/quickstarts)
        * `gcloud components install app-engine-go`
        * `gcloud components install beta cloud-datastore-emulator`
    * [dep](github.com/golang/dep)
        * `go get -u github.com/golang/dep/cmd/dep`

1. Testing in local

```
$ ./setup.sh # exec once
$ ./serve.sh # exec in background
$ ./test.sh
```

2. Testing with Docker

```
$ ./build-in-docker.sh
```

3. Testing with [Circle CI CLI](https://circleci.com/docs/2.0/local-jobs/)

```
$ circleci build
```

## License

Copyright 2017 Mercari, Inc.

Licensed under the MIT License.
