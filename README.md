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


## make the code testable
Once the API of GoGrinder was drafted I started writing tests. Testing in golang takes a little bit more attention than usual. Workaounds for bad design like mocking and monkey-patching are not readly available in Go as they are in other dynamic languages.

Due to my professional emphasis on testing I put a lot of attention on testing. Focus on testing fueled my learning process of the language itself.

### Using Interfaces
http://nathanleclaire.com/blog/2015/10/10/interfaces-and-composition-for-effective-unit-testing-in-golang/
http://nathanleclaire.com/blog/2015/03/09/youre-not-using-this-enough-part-one-go-interfaces/

### Dealing with the golang "time" package (careful, I beat my own drum here!)
https://github.com/finklabs/ttime

### Dealing with the golang "fmt" package (again)
http://stackoverflow.com/questions/34462355/how-to-deal-with-the-fmt-golang-library-package-for-cli-testing/

### Dealing with os.Exit
library code should never os.Exit(1). Instead return from functions using error. Output "error.Error()"" and call "os.Exit(1)" from the calling main function! Funny thing is that this dramatically simplifies the testing of the library code, too.


## testing ui code
http://bendetat.com/karma-and-mocha-for-angular-testing.html
