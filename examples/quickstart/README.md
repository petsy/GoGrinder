Quickstart Guide
====================

Familiarize yourself with the basic GoGrinder concepts.

The [quickstart example in more detail](../../docu/quickstart.md).


## Start the server

In order to demonstrate the use of GoGrinder we need a tiny server that responds to our Http requests.
For every benchmark we need to start a tiny HTTP server to act as our backend.
(we let the server run for 30 seconds and run the test for 20)

$ airbiscuit 30


## Run the GoGrinder test-scenario

First we need to compile the test-scenario into a single executable:

```sh
$ go build -o gogrinder
```

Now run the test-scenario:

```sh
$ ./gogrinder -no-frontend quickstart_loadmodel.json
```
