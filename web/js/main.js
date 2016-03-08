var app = angular.module('gogrinder', [
    'mgcrea.ngStrap',
]);


app.config(function ($httpProvider) {
    $httpProvider.defaults.headers.common['Authorization'] = 'Token 20002cd74d5ce124ae219e739e18956614aab490';
});

// service to start, stop, provide test results
app.controller('MainController', function ($scope, $filter, $modal, $http, TestService, ConfigService) {
    $scope.order = "teststep";
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
        'mtime': "",
        'readConfig': function () {
            $http.get('/config')
                .success(function(data, status, headers, config) {
                    self.loadmodel = JSON.stringify(data["config"], null, 2);
                    self.mtime = data["mtime"];
                })
                .error(function(data, status, headers, config) {
                    // log error
                });
        },
        'saveConfig': function (cb_ok) {
            console.log('save file');
            $http.put('/config', self.loadmodel)
                .success(function(data, status, headers, config) {
                    // TODO add some confirmation
                    console.log("config saved!");
                    self.readConfig();
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
                    if (results[i]['teststep'] === self.results[j]['teststep']) {
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
            $http.get('/statistics?since=' + self.last())
                .success(function (data, status, headers, config) {
                    self.updateResults(data.results);
                    self.running = data.running;
                })
                .error(function (data, status, headers, config) {
                    console.log("Waiting for the user to  restart the test from console...");
                });
        },
        'dataPoller': function () {
            // update results while test is running
            $timeout(function () {
                self.loadResults();
                self.dataPoller();
            }, 1000)
        },
        'start': function () {
            if (self.running) {
                return;
            }
            self.running = true;
            self.results = [];
            $http.post('/test')
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
            $http.delete('/test')
                .success(function (data, status, headers, config) {
                    console.log('test stopped');
                })
                .error(function (data, status, headers, config) {
                    // log error
                });
        },
        'exit': function () {
            $http.delete('/stop')
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
