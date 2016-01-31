# GoGrinder

[![Build Status](https://drone.io/github.com/finklabs/GoGrinder/status.png)](https://drone.io/github.com/finklabs/GoGrinder/latest)
[![GoDoc](https://godoc.org/github.com/finklabs/GoGrinder?status.svg)](https://godoc.org/github.com/finklabs/GoGrinder)
[![License](http://img.shields.io/badge/license-MIT-yellowgreen.svg)](MIT_LICENSE)

Efficient load-generator that integrates in Prometheus for reporting and analysis.

## GoGrinder usage
The GoGrinder is not used directly. You can use GoGrinder as a library to write your load and performance tests. The following sample shows you how: [simple GoGrinder sample]/https://github.com/finklabs/GoGrinder-samples/tree/master/simple)


## Backend development
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


## Frontend development
The frontend is packaged with the executable. To access the frontend start your test and point your browser to:
http://localhost:3030/app/index.html


### adding frontend dependencies using bower

    $ bower install --save angular 
    $ bower install --save angular-ladda
    ...
    $ bower install --save-dev angular-mocks


### testing ui code

* http://www.bradoncode.com/blog/2015/05/19/karma-angularjs-testing/
* http://bendetat.com/karma-and-mocha-for-angular-testing.html

=> we are using mocha, chai, sinon

running the tests unsing karma:

    $ npm test


## Releasing 
Not yet sure what the prevailing strategies for maintaining versions in Golang are. Golang itself has no notation of a package version. I guess this has its origins in the Google development model. As far as I know everyone in Google is on trunk. This approach probably makes a lot of sense within Google - at least I see many of benefits. Obviously for the rest of the world there is no way to avoid dealing with the "dependency hell".
 
One approach that made a lot of sense to me is http://labix.org/gopkg.in. gopkg helps you to maintain multiple versions in one repository:

    gopkg.in/user/pkg.v3 → github.com/user/pkg   (branch or tag v3)


## Where are we now
For this kind of application I believe it is essential to have a core of highest quality. A smaller code base makes this easier to achieve. The Golang concurrency features allow me to keep the code concise and maintainable.  I ran line-counting which came up with 1100 lines of Go code for the core functionality. To me this means two things:
 
1. The goal of a reliable core is in reach and will be achieved soon
2. Golang was the right technology choice for this project

Statistics from 14th January 2016:

    $ cloc --exclude-dir=bower_components,node_modules,web/libs .
        27 text files.
        27 unique files.                              
     11835 files ignored.
    
    http://cloc.sourceforge.net v 1.60  T=0.59 s (33.6 files/s, 3643.5 lines/s)
    -------------------------------------------------------------------------------
    Language                     files          blank        comment           code
    -------------------------------------------------------------------------------
    Go                              10            193            176           1079
    Javascript                       3             67             62            262
    HTML                             2             18             19            153
    CSS                              2              5             10             56
    YAML                             2              8              8             27
    Bourne Shell                     1              6              3             15
    -------------------------------------------------------------------------------
    SUM:                            20            297            278           1592
    -------------------------------------------------------------------------------
