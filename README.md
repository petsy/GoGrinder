GoGrinder
==============

[![Build Status](https://drone.io/github.com/finklabs/GoGrinder/status.png)](https://drone.io/github.com/finklabs/GoGrinder/latest)
[![GoDoc](https://godoc.org/github.com/finklabs/GoGrinder?status.svg)](https://godoc.org/github.com/finklabs/GoGrinder)
[![License](http://img.shields.io/badge/license-MIT-yellowgreen.svg)](MIT_LICENSE)

GoGrinder helps you and your team to check the stability and performance of your code. GoGrinder provides you with an efficient load generator that comes without license restrictions.

Modeling a realistic load profile is easy using loadmodel.json format. Necessary configuration to run the test-scenario with 600 virtual users for half an hour, ramping up 20 users per second:

{"Loadmodel":[
	{"Pacing":0,"Runfor":1800,"Testcase":"01_testcase","Users":300,"Rampup":0.1},
	{"Pacing":0,"Runfor":1800,"Testcase":"02_testcase","Users":300,"Rampup":0.1}
],
"Scenario":"scenario1","ThinkTimeFactor":0,"ThinkTimeVariance":0}

You can simulate from a few to dozens to many hundreds of virtual users using GoGrinder.

For more information, see the

* go doc
* quickstart


## Installation

Compile your test-scenarios into a single executable. Usually we keep test-scenarios in the gogrinder.go source file. You do not need to install a compiler. Simply use Docker to run the compiler:

$ docker run bla bla TODO -o gogrinder

This compiles your testscenario including everything that is necessary into the gogrinder executable. Just put the gogrinder executable and loadmodel.json wherever you want to run the test.

$ ./gogrinder yourcode_loadmodel.json

Alternatively, if you have Go installed you can also use this compiler:

$ go build -o gogrinder


## Examples

* xmlcowboys - showcase demonstrates how to read data from XML files and use it as http requests.

* supercars - a more complete sample uses redis to exchange data between virtual users.


## Grafana and Prometheus

Use Grafana and Prometheus to visualize metrics and test results. Don't worry this won't become a headache! We use Docker to setup these tools in minutes and we give you the instructions to do the same.

use LICEcap to make an visualization animation


## Key Features

The key features of GoGrinder are:

* **Unlimited-Load**: GoGrinder does not limit the number of virtual users or transactions per second or anything. No dual-licensing, nothing. All we ask from you is to contribute back and to help us promote GoGrinder. 

* **Efficient**: run hundreds of virtual users

* **learning-curve**: GoGrinder is build in a way to help you setup tests quickly and easily. In this way will supports your team to developing skills from running first performance tests into more advance performance management. 

* **Workflow**: GoGrinder is made so it can be run fully automated to support the continuous-performance usecase. The built-in web console (http://localhost:3030/app). Makes it easy to start and debug your performance test-scenarios right from the beginning.

* **Deployment**: GoGrinder test-scenarios are compiled into single executables. You need only the executable plus the loadmodel.json file to deploy your test.

* **Visualization**: GoGrinder ... best of breed monitoring and visualization tools

* **Extensibility**: Many enterprise applications are using proprietary or exotic client server protocols. GoGrinder is extensible so you can add support for these protocols yourself.

* **Docker**: GoGrinder encourages the use of Docker. Applications can be containerized
  to make deployments and performance testing easier without changing the developer
  workflow. GoGrinder recommends to use Docker to ease test development and workflow.

* **Quality**: GoGrinder itself has gone through many testing cycles to make sure you get the best experience possible. GoGrinder is covered by a complete unit and integration test suite. GoGrinder performance is checked frequently using runtime/pprof.


## Getting help

We think our mode to supporting you is pretty common so please excuse captain-obvious.

In case you are stuck or need help to solve your problem, please follow this checklist:

* look in documentation and examples for a solution
* for Go related information the std. library docu is a good starting point TODO
* if you are certain you caught a bug in GoGrinder look for / open an issue
* search for a solution to your problem on stackexchange
* if you can not find a solution please open a new question. Please make sure to tag your question with the following tags: "Performance", "GoGrinder" and "Go"

We monitor for GoGrinder related questions and try to help you asap. Please forgive but we all have day-jobs and it might take us a day or two to answer your question.

TODO check fair use


## Developing GoGrinder

If you wish to work on GoGrinder itself or any of its built-in systems,
you'll first need [Go](https://www.golang.org) installed on your
machine (version 1.4+ is *required*).

For local dev first make sure Go is properly installed, including setting up a
[GOPATH](https://golang.org/doc/code.html#GOPATH).

I you want to contribute your code please take a look into the contribute... TODO
