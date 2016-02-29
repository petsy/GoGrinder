Using GoGrinder on Windows TM
==================================

Everything is basically the same on the Windows TM environment. However subtle differences exist. For that reason we decided to walk you trough an example step by step on how to do things on a windows environment...


## Installation

```sh
$ export GOOS=windows
$ export GOARCH=386
$ go build -o gogrinder.exe
```

Compile your test-scenarios into a single executable. Usually we keep test-scenarios in the gogrinder.go source file. You do not need to install a compiler. Simply use Docker to run the compiler:

```sh
$ docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e GOOS=windows -e GOARCH=386 golang:1.5-cross go build -o gogrinder
```

This results in a gogrinder.exe binary.

You can also use this command to **cross-compile** the test-scenario for Windows TM targets on a Linux or Mac environment!


## Running your tests

After you compiled your test-scenario for the Windows TM platform you are ready to rock! Just put the gogrinder.exe and loadmodel.json on the machine where you want to run the test.

```sh
$ gogrinder.exe yourcode_loadmodel.json
```
