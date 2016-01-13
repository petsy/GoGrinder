/**
 * Created by mark on 1/12/16.
 */


describe('TestService', function () {
    var test, httpBackend;

    beforeEach(module('gogrinder'));

    beforeEach(inject(function (TestService, $httpBackend, $timeout) {
        test = TestService;
        httpBackend = $httpBackend;
        timeout = $timeout;
    }));

    describe('exit()', function () {
        it('should call DELETE on REST service method /stop', function () {
            httpBackend
                .expect('DELETE', 'http://localhost:3000/stop')
                .respond(200, '');

            test.exit();

            httpBackend.flush();
        });
    });


    describe('updateResults()', function () {
        it('should update empty initial set', function () {
            test.results = [];
            var b = [
                {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
                {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
                {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            ];
            test.updateResults(b);
            assert.equal(3, test.results.length, "Expected 3 elements after update!");
        });

        it('should update with empty set', function () {
            test.results = [
                {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
                {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
                {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            ];
            test.updateResults([]);
            assert.equal(3, test.results.length, "Expected 3 elements after update!");
        });

        it('should append disjunct sets', function () {
            test.results = [
                {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
            ];
            var b = [
                {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
                {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            ];
            test.updateResults(b);
            assert.equal(3, test.results.length, "Expected 3 elements after update!");
        });

        it('should handle intersecting sets', function () {
            test.results = [
                {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
                {"testcase": "02_01_teststep", "avg": 0, "min": 0, "max": 0, "count": 0},
            ];
            var b = [
                {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
                {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            ];
            var expected = [
                {"testcase": "01_01_teststep", "avg": 100222663, "min": 100151989, "max": 100303219, "count": 18},
                {"testcase": "02_01_teststep", "avg": 200227442, "min": 200133951, "max": 200279875, "count": 9},
                {"testcase": "03_01_teststep", "avg": 300230474, "min": 300194438, "max": 300263493, "count": 6}
            ]
            test.updateResults(b);
            assert.equal(3, test.results.length, "Expected 3 elements after update!");
            assert.deepEqual(expected, test.results, "Expected update to merge intersecting sets!");
        });
    });


    describe('last()', function () {
        it('should come up with empty timestamp for empty result set', function () {
            test.results = []
            assert.equal("", test.last(), "Expected '' as last timestamp!");
        });
        it('should come up with the last timestamp', function () {
            test.results = [{"last": "0815"}]
            assert.equal("0815", test.last(), "Expected '0815' as last timestamp!");
        });
        it('should come up with the last timestamp from multiple results', function () {
            test.results = [{"last": "0816"}, {"last": "0815"}]
            assert.equal("0816", test.last(), "Expected '0816' as last timestamp!");
        });
    });


    describe('loadResults()', function () {
        it('should load the first result set', function () {
            test.results = []

            httpBackend
                .expect('GET', 'http://localhost:3000/statistics?since=')
                .respond(200, {"results": [{"testcase": "01", "last": "0816"}], "running": true});
            test.loadResults();
            httpBackend.flush();

            assert.deepEqual([{"testcase": "01", "last": "0816"}],
                test.results, "Expected loadResults() to update the result set!");
        });

        it('should update results', function () {
            test.results = [{"testcase": "01", "last": "0815"}]

            httpBackend
                .expect('GET', 'http://localhost:3000/statistics?since=0815')
                .respond(200, {"results": [{"testcase": "01", "last": "0816"}], "running": true});
            test.loadResults();
            httpBackend.flush();

            assert.deepEqual([{"testcase": "01", "last": "0816"}],
                test.results, "Expected loadResults() to update the result set!");
        });
    });


    describe('dataPoller()', function () {
        it('should run until test stops', function () {
            test.results = [{"testcase": "01", "last": "0815"}];
            // note: dataPoller runs automatically with TestService

            httpBackend
                .expect('GET', 'http://localhost:3000/statistics?since=0815')
                .respond(200, {"results": [{"testcase": "01", "last": "0816"}], "running": false});
            timeout.flush();
            httpBackend.flush();

            // note: now the test ist stopped ("running": false)
            // dataPoller has exited (test would complain about unexpected requests!)
            timeout.flush();
        });
    });


});
