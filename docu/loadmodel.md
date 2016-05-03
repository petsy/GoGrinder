Modeling a realistic load profile is easy using loadmodel.json format. The following sample shows the necessary configuration to run the test-scenario with 600 virtual users for half an hour, at start it ramps-up 20 users per second:

```javascript
{"Loadmodel":[
	{"Pacing":0,"Runfor":1800,"Testcase":"01_testcase","Users":300,"Rampup":0.1},
	{"Pacing":0,"Runfor":1800,"Testcase":"02_testcase","Users":300,"Rampup":0.1}
],
"Scenario":"scenario1","ThinkTimeFactor":0,"ThinkTimeVariance":0}
```

