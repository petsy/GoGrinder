var app = angular.module('gogrinder', [
    'ngResource',
    'infinite-scroll',
    'angularSpinner',
    'jcs-autoValidate',
    'angular-ladda',
    'mgcrea.ngStrap',
    'toaster',
    'ngAnimate'
]);


app.config(function ($httpProvider, $resourceProvider, laddaProvider) {
    $httpProvider.defaults.headers.common['Authorization'] = 'Token 20002cd74d5ce124ae219e739e18956614aab490';
    $resourceProvider.defaults.stripTrailingSlashes = false;
    laddaProvider.setOption({
        style: 'expand-right'
    });
});

// service to start, stop, provide test results
app.controller('MainController', function ($scope, $filter, TestService) {
    $scope.order = "testcase";
    $scope.reverse = false;
    $scope.test = TestService;

    //$scope.showCreateModal = function () {
    //	$scope.contacts.selectedLoadmodel = {};
    //	$scope.createModal = $modal({
    //		scope: $scope,
    //		template: 'templates/modal.create.tpl.html',
    //		show: true
    //	})
    //};

});


// service to start, stop, provide test results
app.service('TestService', function($http, $timeout) {
    var self = {
        'running': true,
        'results': [],
        'last': function() {
            var last = null;
            angular.forEach(self.results, function(value, key) {
                if (last == null || last < value['last']) { last = value['last'] }
            });
            console.log(last)
            return last;
        },
        'updateResults': function(results) {
            // no merge for javascript arrays so we implement this, too
            for(var i=0; i<results.length; i++) {
                // loop over results
                var found = false;
                for(var j=0; j<self.results.length; j++) {
                    // loop over self.results
                    if(results[i]['testcase'] === self.results[j]['testcase']) {
                        //update
                        self.results[j] = results[i];
                        found = true;
                        break;
                    }
                }
                // append
                if (!found) {self.results.push(results[i])}
            }
        },
        'loadResults': function() {
            //self.results = [
            //    {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
            //    {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
            //    {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            //];

            $http.get('http://localhost:3000/statistics?since=' + self.last())
                .success(function(data, status, headers, config) {
                    self.updateResults(data.results);
                    self.running = data.running;
                    console.log(self.running);
                })
                .error(function(data, status, headers, config) {
                // log error
                });
        },
        'dataPoller': function() {
            // update results while test is running
            $timeout(function() {
                self.loadResults();
                console.log('updated');
                console.log(self.running);
                if (self.running) { self.dataPoller(); }
            }, 1000)

        },
        'start': function() {
            if (self.running) { return; }
            self.running = true;
            self.results = [];
            self.dataPoller();
            $http.post('http://localhost:3000/test')
                .success(function(data, status, headers, config) {
                    console.log('test started');
                })
                .error(function(data, status, headers, config) {
                    // log error
                });
        },
        'stop': function() {
            if (!self.running) { return; }
            $http.delete('http://localhost:3000/test')
                .success(function(data, status, headers, config) {
                    console.log('test stoped');
                })
                .error(function(data, status, headers, config) {
                    // log error
                });
        },
        'exit': function() {
            $http.delete('http://localhost:3000/stop')
                .success(function(data, status, headers, config) {
                    console.log('webserver stoped');
                })
                .error(function(data, status, headers, config) {
                    // log error
                });
        },
    };

    self.dataPoller();
    return self;
});
