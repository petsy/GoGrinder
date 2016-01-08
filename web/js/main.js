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


app.config(function ($httpProvider, $resourceProvider, laddaProvider, $datepickerProvider) {
	$httpProvider.defaults.headers.common['Authorization'] = 'Token 20002cd74d5ce124ae219e739e18956614aab490';
	$resourceProvider.defaults.stripTrailingSlashes = false;
	laddaProvider.setOption({
		style: 'expand-right'
	});
	angular.extend($datepickerProvider.defaults, {
		dateFormat: 'd/M/yyyy',
		autoclose: true
	});
});

// service to start, stop, provide test results
app.controller('MainController', function ($scope, $modal, TestService) {
	$scope.order = "testcase";
	$scope.test = TestService;


	$scope.start = function () {
		console.log("start test");
		$scope.test.running = !$scope.test.running
		//$scope.execution.createContact($scope.contacts.selectedPerson)
		//    .then(function () {
		//        $state.go("list");
		//    })
	};
	$scope.stop = function () {
		console.log("stop test");
		$scope.test.running = !$scope.test.running
		//$scope.execution.createContact($scope.contacts.selectedPerson)
		//    .then(function () {
		//        $state.go("list");
		//    })
	};


	//$scope.loadMore = function () {
	//	console.log("Load More!!!");
	//	$scope.test.loadMore();
	//};
    //
	//$scope.showCreateModal = function () {
	//	$scope.contacts.selectedLoadmodel = {};
	//	$scope.createModal = $modal({
	//		scope: $scope,
	//		template: 'templates/modal.create.tpl.html',
	//		show: true
	//	})
	//};

	$scope.$watch('order', function (newVal, oldVal) {
		if (angular.isDefined(newVal)) {
			$scope.test.doOrder(newVal);
		}
	})

});


// service to start, stop, provide test results
app.service('TestService', function () {
 	var self = {
		// running, results, last that is all
		'running': false,
		'results': [],
		'last': null,
 		'doOrder': function (order) {
 			//self.hasMore = true;
 			//self.page = 1;
 			//self.persons = [];
 			self.ordering = order;
 			self.loadResults();
 		},
 		'loadResults': function () {
			self.results = [
			{"name":"01_01_teststep","avg":100222663,"min":100151989,"max":100303219,"count":18},
			{"name":"02_01_teststep","avg":200227442,"min":200133951,"max":200279875,"count":9},
			{"name":"03_01_teststep","avg":300230474,"min":300194438,"max":300263493,"count":6}
			];

 		},
	};

	self.loadResults();
	return self;
});
