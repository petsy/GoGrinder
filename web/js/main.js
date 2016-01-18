var app = angular.module('gogrinder', [
    //'ngResource',
    //'infinite-scroll',
    //'angularSpinner',
    //'jcs-autoValidate',
    //'angular-ladda',
    'mgcrea.ngStrap',
    //'toaster',
    //'ngAnimate'
]);


app.config(function ($httpProvider) {
    //laddaProvider
    $httpProvider.defaults.headers.common['Authorization'] = 'Token 20002cd74d5ce124ae219e739e18956614aab490';
    //$resourceProvider.defaults.stripTrailingSlashes = false;
    //laddaProvider.setOption({
    //    style: 'expand-right'
    //});
});

// service to start, stop, provide test results
app.controller('MainController', function ($scope, $filter, $modal, $http, TestService, ConfigService) {
    $scope.order = "testcase";
    $scope.reverse = false;
    $scope.test = TestService;
    $scope.config = ConfigService;
    $scope.revision = "";

    $scope.editConfig = function() {
        $scope.editModal = $modal({
            scope: $scope,
            template: 'templates/config.html',
            show: true
        })
    };
    $scope.showAbout = function() {
        $http.get('revision.txt')
            .success(function (data) {
                $scope.revision = data;
            });
        $scope.aboutModal = $modal({
            scope: $scope,
            template: 'templates/about.html',
            show: true
        })
    };
    $scope.save = function () {
        $scope.config.saveConfig(function(){
            $scope.editModal.hide();
        });
    };

});


// service to start, stop, provide test results
app.service('ConfigService', function ($http, $timeout) {
    var self = {
        'loadmodel': "",
        'readConfig': function () {
            $http.get('http://localhost:3000/config')
                .success(function(data, status, headers, config) {
                    self.loadmodel = JSON.stringify(data, null, 2);
                })
                .error(function(data, status, headers, config) {
                    // log error
                });
        },
        'saveConfig': function (cb_ok) {
            console.log('save file');
            $http.put('http://localhost:3000/config', self.loadmodel)
                .success(function(data, status, headers, config) {
                    // TODO add some confirmation
                    console.log("config saved!");
                    cb_ok();
                })
                .error(function(data, status, headers, config) {
                    // log error
                    // TODO user feedback on validation errors etc.
                });
        }
    };

    self.readConfig();
    return self;
});


// service to start, stop, provide test results
app.service('TestService', function ($http, $timeout) {
    var self = {
        'running': true,
        'results': [],
        'last': function () {
            var last = '';
            angular.forEach(self.results, function (value, key) {
                if (last == null || last < value['last']) {
                    last = value['last']
                }
            });
            return last;
        },
        'updateResults': function (results) {
            // no merge for javascript arrays so we implement this, too
            for (var i = 0; i < results.length; i++) {
                // loop over results
                var found = false;
                for (var j = 0; j < self.results.length; j++) {
                    // loop over self.results
                    if (results[i]['testcase'] === self.results[j]['testcase']) {
                        //update
                        self.results[j] = results[i];
                        found = true;
                        break;
                    }
                }
                // append
                if (!found) {
                    self.results.push(results[i])
                }
            }
        },
        'loadResults': function () {
            $http.get('http://localhost:3000/statistics?since=' + self.last())
                .success(function (data, status, headers, config) {
                    self.updateResults(data.results);
                    self.running = data.running;
                })
                .error(function (data, status, headers, config) {
                    // log error
                });
        },
        'dataPoller': function () {
            // update results while test is running
            $timeout(function () {
                if (self.running) {
                    self.loadResults();
                }
                if (self.running) {
                    self.dataPoller();
                }
            }, 1000)

        },
        'start': function () {
            if (self.running) {
                return;
            }
            self.running = true;
            self.results = [];
            self.dataPoller();
            $http.post('http://localhost:3000/test')
                .success(function (data, status, headers, config) {
                    console.log('test started');
                })
                .error(function (data, status, headers, config) {
                    // log error
                });
        },
        'stop': function () {
            if (!self.running) {
                return;
            }
            $http.delete('http://localhost:3000/test')
                .success(function (data, status, headers, config) {
                    console.log('test stopped');
                })
                .error(function (data, status, headers, config) {
                    // log error
                });
        },
        'exit': function () {
            $http.delete('http://localhost:3000/stop')
                .success(function (data, status, headers, config) {
                    console.log('webserver stopped');
                })
                .error(function (data, status, headers, config) {
                    // log error
                });
        }
    };

    self.dataPoller();
    return self;
});
