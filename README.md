# GoGrinder

[![Build Status](https://drone.io/github.com/finklabs/GoGrinder/status.png)](https://drone.io/github.com/finklabs/GoGrinder/latest)
[![License](http://img.shields.io/badge/license-MIT-yellowgreen.svg)](MIT_LICENSE)

Efficient load-generator that integrates in Prometheus for reporting and analysis.

## GoGrinder usage
The GoGrinder is not used directly. You can use GoGrinder as a library to write your load and performance tests. The following sample shows you how:
https://github.com/finklabs/GoGrinder-samples/tree/master/simple


## backend development
The reminder of the document contains information is intended for developers who work on the backend and frontend of GoGrinder.

### Run the go package docu (for offline use)
$ godoc -http=:6060 &


### Embedd the single page app
$ rice embed-syso


### Run the tests with coverage report
$ gocov test | gocov report


### build the package
$ ./build.sh

build the package:
$ go build

install into pkg folder:
$ go install


## frontend development
The frontend is packaged with the executable. To access the frontend start your test and point your browser to:
http://localhost:3000/app/index.html


### adding frontend dependencies using bower
$ bower install --save angular 
$ bower install --save angular-ladda
...
$ bower install --save-dev angular-mocks


### testing ui code
http://www.bradoncode.com/blog/2015/05/19/karma-angularjs-testing/
http://bendetat.com/karma-and-mocha-for-angular-testing.html
=> we are using mocha, chai, sinon

running the tests unsing karma:
$ npm test
