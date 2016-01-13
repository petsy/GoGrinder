#!/bin/bash

libs="../web/libs/"

# copy all necessary files over to web/libs/
cd bower_components
cp --parents angular/angular.min.js $libs
cp --parents bootstrap/dist/css/bootstrap.min.css $libs
cp --parents bootstrap-additions/dist/bootstrap-additions.min.css $libs
cp --parents font-awesome/css/font-awesome.min.css $libs
cp --parents font-awesome/fonts/fontawesome-webfont.woff2 $libs
cp --parents font-awesome/fonts/fontawesome-webfont.woff $libs
cp --parents font-awesome/fonts/fontawesome-webfont.ttf $libs
cp --parents angular-strap/dist/angular-strap.min.js $libs
cp --parents angular-strap/dist/angular-strap.tpl.min.js $libs
cd ..


# embedd the frontend
rice embed-syso


# build the go package
go build
