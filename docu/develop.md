Developing GoGrinder and Plugins
=========================================

This documentation contains information relevant to maintaining developing GoGrinder and Plugins

* setup a development environment
* testing
* release


## Backend development

### Run the go package docu (for offline use)

```sh
$ godoc -http=:6060 &
```


### Run the tests with coverage report

```sh
$ gocov test | gocov report
```


### build the package

```sh
$ ./build.sh
```


## Frontend development

The frontend is packaged with the executable. To access the frontend start your test and point your browser to:
http://localhost:3030/app/index.html


### adding frontend dependencies using bower

```sh
$ bower install --save angular 
$ bower install --save angular-ladda
...
$ bower install --save-dev angular-mocks
```


### testing ui code

* http://www.bradoncode.com/blog/2015/05/19/karma-angularjs-testing/
* http://bendetat.com/karma-and-mocha-for-angular-testing.html

=> we are using mocha, chai, sinon

running the tests unsing karma:

```sh
$ npm test
```


## Releasing 

We are not yet complete sure what the prevailing strategies for maintaining versions in Golang are. Golang itself has no notation of a package version. I guess this has its origins in the Google development model. As far as I know everyone in Google is on trunk. This approach probably makes a lot of sense within Google - at least I see many of benefits. Obviously for the rest of the world there is no way to avoid dealing with the "dependency hell".
 
One approach that made a lot of sense to me is http://labix.org/gopkg.in. gopkg helps you to maintain multiple versions in one repository:

gopkg.in/user/pkg.v3 → github.com/user/pkg   (branch or tag v3)
