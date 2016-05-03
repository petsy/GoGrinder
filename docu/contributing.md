# Contributing to GoGrinder

We started at the end of 2015 and consequently there are many areas where you can help. This document contains information for you in case you want to contribute to the GoGrinder. The CLA - Contributing License Agreement applies. 

## Contributing code

We had some significant API changes recently. So currently it looks good and stable, but working with more plugins we might face new issues that require API changes. We need to deal with that, too.

If you submit a pull-request, then please make sure that it contains a file yourname.md in docu/contributors. The file should contain something like the following: "I (insert first name, last name, email address) have read and understood the contributors license agreement (docu/contributors/CLA.md) and fully agree to it." (Insert location and date).

If you contribute code there are some important things: test coverage, test coverage, test coverage... Did I mention test coverage?

In order to make your contribution useful to others you should at least add some basic documentation and if possible a small sample. For non Go samples we use Docker to simplify things. 

The http plugin is model we used for designing the GoGrinder API. It can be used as a model to develop other plugins, too. 

## Areas we are looking for contributors 

If you need an idea for a plugin, then I think the most important plugins to work on are:

* IMAP
* Mysql
* mgo
* Protobuf (I think there are already some Prometheus metrics for that)
* NoSQL (Redis, ELS, Solr)

## Feedback

We are still at a very early stage so we very much appreciate and need your feedback. If you are using GoGrinder to test your website or your application we would love to hear from you. We hope you do not run into any trouble. But if you do - we want to hear it. Of course we are going to help you!

The best way to get in touch with us is via https://github.com/finklabs/GoGrinder use issues and pull-requests for this.

If you like GoGrinder and find it useful, than please spread the word! Thank you!



