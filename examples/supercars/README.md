# Performance Testing the Supercars application

In many performance tests it is necessary to exchange testdata (test state) between multiple testcases / virtual users. Or to reserve test resources like user accounts. In this howto I use Redis to exchange supercar ids between multiple testcases. This solution could be applied for a performance test of a more realistic application with a workflow. Careful - for testing the supercars application this is not the simplest solution possible. You could just chain the test steps together and get rid of Redis (the jmeter performance test implementation uses this approach).


## Running the Supercars application

The "preferred" way of running the supercars application is via Docker. Of cause there are many other ways depending on your OS, setup and skill set. We figured that the easiest and most portable way is via Docker. On most OSs Docker provides you close to bare metal performance so we should be fine using this for performance testing.

$ docker run -p 8000:8000 finklabs/supercars

now direct your browser to:
http://localhost:8000/app/


## Running Redis

Sure it would be fun to rewrite Redis or at least to use a Redis clone which is implemented in Golang. For me Redis works great so for now I keep using it. If Redis is a red flag for you then it should be fairly easy to refactor the samples so employ your favorite tools. Enough chatter - lets get started...

$ docker run -p 6379:6379 redis


## TODO start node exporter, prometheus and grafana

...

get us some nice graphs, too


## Running the test

$ go run gogrinder.go supercars_loadmodel.json

Now point your browser to the following url: http://localhost:3030/app/


## First preliminary results

01_01_supercars_list, 12.360785, 3.290222, 25.568050, 301, 0
02_01_supercars_read, 3.569725, 1.196538, 8.187192, 302, 0
03_01_supercars_create, 3.588597, 1.305598, 7.471673, 301, 0
04_01_supercars_update, 3.431132, 1.068884, 8.576453, 301, 0
05_01_supercars_delete, 3.817503, 1.088714, 9.443368, 300, 0
