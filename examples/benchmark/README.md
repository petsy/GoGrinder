# Using different tools for benchmarking

Benchmarking tools:

* vegeta
* jmeter
* grinder
* ab
* gogrinder

Disclaimer: Please note that most of the tools have a completely different focus. Therefore it is best to select a suitable tool based on your specific needs. Objective of the benchmark is to show that all tools get the job done.


## Approach

For every benchmark we need to start a tiny HTTP server to act as our backend.
(we let the server run for 30 seconds and run the test for 20)

$ airbiscuit 30


## Vegeta

$ go get github.com/tsenart/vegeta

$ ./vegeta_get


## ApacheBench (AB)

$ sudo apt-get install apache2-util

$ ab -t 20 -c 1000 http://localhost:3001/get_stuff


## Jmeter

$ ./apache-jmeter-2.13/bin/jmeter.sh -t jmeter_test_plan.jmx


## GoGrinder

$ ./gogrinder -no-frontend gogrind_loadmodel.json


## Benchmark testobject concept

some research on request size:
http://www.websiteoptimization.com/speed/tweak/average-web-page/

GET size 1830K ~100 Objects => avg 18K
