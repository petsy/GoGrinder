# GoGrinder

[![Build Status](https://travis-ci.org/finklabs/GoGrinder.svg?branch=master)](https://travis-ci.org/finklabs/GoGrinder)
[![License](http://img.shields.io/badge/license-MIT-yellowgreen.svg)](MIT_LICENSE)

Efficient load-generator that integrates in Prometheus for reporting and analysis.


## Run the go package docu (for offline use)

$ godoc -http=:6060 &


## Embedd the single page app

$ rice embed-syso


## Run the tests with coverage report

$ gocov test | gocov report


## build the package

$ go build


## testing ui code
http://www.bradoncode.com/blog/2015/05/19/karma-angularjs-testing/
http://bendetat.com/karma-and-mocha-for-angular-testing.html
=> we are using mocha, chai, sinon

running the tests:
$ npm test

## adding dependencies using bower
$ bower install --save angular 
$ bower install --save angular-ladda
...
$ bower install --save-dev angular-mocks
