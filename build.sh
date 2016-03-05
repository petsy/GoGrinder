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


# add the revision hash
# note: by definition the revision is always one commit behind so it should be done in a sep. commit step
# this could be done using .gitattributes and a script (this requires a build after checkout, too):
# http://stackoverflow.com/questions/16765461/git-add-revision-number-to-file-on-commit
git rev-parse --short HEAD > web/revision.txt

# embedd the frontend (go.rice turned out to not work on Windows TM)
#rice embed-syso
#./embedsyso

# now we are using https://github.com/elazarl/go-bindata-assetfs
# TODO: wait for https://github.com/elazarl/go-bindata-assetfs/issues/34
# go-bindata-assetfs -pkg gogrinder -o gogrinder/assetfs.go web/...
go-bindata-assetfs -pkg gogrinder web/...
mv bindata_assetfs.go gogrinder/assetfs.go
