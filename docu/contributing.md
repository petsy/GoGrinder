# Contributing to GoGrinder

This document contains information for you in case you want to help out.

# Areas we are urgently looking for contributions

We started in 2015 and consequently there are many areas where you can help out.


## Feedback

We are still at a very early stage so we need your feedback. If you are using GoGrinder to test your website or your application we want to hear from you. We hope you do not run into any trouble. But if so we want to hear it. Of course we are going to help you!

Currently there is no mailing list. But we will get one soon. So currently the best way to get in touch with us is via https://github.com/finklabs/GoGrinder use issues and pull-requests for this.

If you like GoGrinder & find it useful than please spread the word!


## Contributing code

We had some significant API changes recently. So currently it looks good and stable, but working with more plugins we might face new issues that require API changes. We need to deal with that, too

If you submit a pull-request (https://help.github.com/articles/using-pull-requests/), then please make sure that it contains a file <your_name>.md in docu/contributors. The file should contain something like the following: "I <first> <last>, <email> have read, understand and fully agree to the contributors license agreement (docu/contributors/CLA.md)"

If you contribute code there are some important things: test coverage, test coverage, test coverage... Did I say test coverage?

In order to make your contribution useful to others you should at least add some basic documentation and if possible a small sample. For non Go samples we use Docker to simplify things. 

The http plugin is model we used for designing the GoGrinder API. It can be used as a model to develop other plugins, too. 

If you need an idea for a plugin, then I think the most important plugins to work on are:

* IMAP
* Mysql
* mgo
* Protobuf (I think there are already some Prometheus metrics for that)
* NoSQL (Redis, ELS, Solr)