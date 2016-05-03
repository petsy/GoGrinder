GoGrinder
==============

[![Build Status](https://drone.io/github.com/finklabs/GoGrinder/status.png)](https://drone.io/github.com/finklabs/GoGrinder/latest)
[![GoDoc](https://godoc.org/github.com/finklabs/GoGrinder?status.svg)](https://godoc.org/github.com/finklabs/GoGrinder/gogrinder)
[![License](http://img.shields.io/badge/license-MIT-yellowgreen.svg)](LICENSE/)

GoGrinder is a performance test tool designed to help test engineers check the stability and performance of their code and pinpoint problems of their applications under stress. 

It simulates concurrent user activity and monitors system resources and performance behaviour - from a handful to hundreds of virtual users. GoGrinder provides an efficient load generator without license restrictions. It is written in GoLang, and integrates in Prometheus for reporting and analysis. 

## Origins, Need & Benefits

Is your application, server, or service delivering the appropriate speed of need? How do you know? Are you 100-percent certain that your latest feature hasn’t triggered a performance degradation or memory leak? There's only one way to verify - and that's by regularly checking the performance of your app. A performance test tool allows you to easily identify website and application performance bottlenecks and points of failure. Analyzing and optimizing digital business performance—in real time is more than monitoring, it’s true Application Intelligence. With The GoGrinder you can transform application and operational insights into a competitive advantage.

As an engineer I have more than ten years experience in the field of performance engineering of ecommerce and banking websites. Here solid tooling is one of the most important aspects. I worked with all relevant proprietary and open source tools. I favored working with The Grinder (http://grinder.sourceforge.net/).

However. The Grinder today is 15+ years old. It has been designed and built for a pre-cloud/pre-container world. Things are different now. I gave The Grinder a complete overhaul, simplifying it, while keeping its conclusive approach to load testing intact. 

The result is the GoGrinder.

The GoGrinder is installed and ready to use within minutes (simple installation with just one file). It is straightforward, smart, adaptive, customizable, highly flexible and open-source. It can be used in the cloud (default) and it can test Single Page Applications (SPA). The GoGrinder is automated - providing consistent, repeatable and traceable results. No fuss, no frills, open-source.

## The key features of the GoGrinder are:

* **Unlimited-Load**: GoGrinder does not limit the number of virtual users or transactions per second or anything. No dual-licensing, nothing. All we ask from you is to contribute back and to help us promote GoGrinder. 
* **Efficient**: run hundreds of virtual users
* **Learning-curve**: GoGrinder is build in a way to help you setup tests quickly and easily. In this way will supports your team to developing skills from running first performance tests into more advance performance management. 
* **Workflow**: GoGrinder is made so it can be run fully automated to support the continuous-performance usecase. The built-in web console (http://localhost:3030/app). Makes it easy to start and debug your performance test-scenarios right from the beginning.
* **Deployment**: GoGrinder test-scenarios are compiled into single executables. You need only the executable plus the loadmodel.json file to deploy your test.
* **Visualization**: GoGrinder ... best of breed monitoring and visualization tools
* **Extensibility**: Many enterprise applications are using proprietary or exotic client server protocols. GoGrinder is extensible so you can add support for these protocols yourself.
* **Docker**: GoGrinder encourages the use of Docker. Applications can be containerized
  to make deployments and performance testing easier without changing the developer
  workflow. GoGrinder recommends to use Docker to ease test development and workflow.
* **Quality**: GoGrinder itself has gone through many testing cycles to make sure you get the best experience possible. GoGrinder is covered by a complete unit and integration test suite. GoGrinder performance is checked frequently using runtime/pprof.

For more information, see the

* [GoDoc Documentation](https://godoc.org/github.com/finklabs/GoGrinder/gogrinder)

## Getting started

We are currently working on a Quick Start Guide. Installation instructions see below. 

## Installation

Compile your test-scenarios into a single executable. Usually we keep test-scenarios in the gogrinder.go source file. You do not need to install a compiler. Simply use Docker to run the compiler:

```sh
$ docker run bla bla TODO -o gogrinder
```

This compiles your test-scenario including everything that is necessary into the gogrinder executable. Just put the gogrinder executable and loadmodel.json on a Linux machine where you want to run the test.

```sh
$ ./gogrinder yourcode_loadmodel.json
```

Alternatively, if you have Go installed you can also use this compiler:

```sh
$ go build -o gogrinder
```
## Examples

We provide some examples to get you started using GoGrinder to test performance of your code:

* [**quickstart**](examples/quickstart/) - simple walkthrough on how to use GoGrinder to test the performance of your code.
* [**xmlcowboys**](examples/xmlcowboys/) - showcase on how to read data from massive XML files and send it to a webservice.
* [**cookies**](examples/cookies/) - demonstrate how to use a login-form and cookies.
* [**supercars**](examples/supercars/) - a more advanced but complete http example including monitoring. We use redis to exchange data between virtual users.
* [**simple**](examples/simple/) - an oversimplified sample used for testing and to demonstrate the core concepts of GoGrinder.

## Grafana and Prometheus

Use Grafana and Prometheus to visualize metrics and test results. We use Docker to setup these tools. 

* [Grafana on Github](https://github.com/grafana/grafana)
* [Grafana Website](http://grafana.org/)

* [Prometheus on Github](https://github.com/prometheus/prometheus)
* [Prometheus Website](https://prometheus.io/)

## Getting help

In case you are stuck or need help to solve your problem, please follow this checklist:

* look in [documentation](docu/) and [examples](examples/) for a solution
* for Go related information the [std. library docu](https://golang.org/pkg/net/http/) is a good starting point
* if you feel you caught a bug in GoGrinder look for [related issues](https://github.com/finklabs/GoGrinder/issues)
* search for a solution to your problem on [StackOverflow](http://stackoverflow.com/questions/tagged/gogrinder)
* if you can not find a solution on StackOverflow we suggest you ask a new question. Please make sure to tag your question with the following tags: "Performance", "GoGrinder" and "Go"

We monitor the [GoGrinder](http://stackoverflow.com/questions/tagged/gogrinder) tag on StackOverflow and try to answer asap. We all have day-jobs, please understand that it might take us a day or so to answer your question.

## Developing GoGrinder

If you wish to work on GoGrinder itself or any of its built-in systems,
you'll first need [Go](https://www.golang.org) installed on your
machine (version 1.4+ is *required*).

For local dev first make sure Go is properly installed, including setting up a
[GOPATH](https://golang.org/doc/code.html#GOPATH).

## Contributing to GoGrinder

If you would like to contribute your code please take a look into [how to contribute](docu/contributing.md).

## Licensing

GoGrinder is licesed under the [MIT license (2015)](license). 



