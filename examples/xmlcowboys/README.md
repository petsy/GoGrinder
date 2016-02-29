# The xmlcowboys sample

I wrote the xmlcowboys sample to demonstrate how we can read testdata from xml files (sometimes this is still necessary).

To run the xmlcowboys sample we need to start a tiny HTTP server to act as our backend.
(we let the server run for 30 seconds and run the test for 20)

$ airbiscuit 30


## Running the test

$ go run gogrinder.go -no-frontend xmlcowboys_loadmodel.json

01_01_xmlcowboys_post, 501.677008, 501.285029, 501.847503, 7, 0


## Improvements for the xmlcowboys sample

If you look into the event-log.txt file you will notice that airbiscuit responds with status code 400 since it expects post requests bigger than 2000 bytes. Probably the best way to fix this is to increase the payload size in the bang_x.xml files.
Another imperfection is that airbiscuit is not able to process XML files (yet) so we use RawPost.
