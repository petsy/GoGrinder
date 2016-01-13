/**
 * Created by mark on 1/12/16.
 */


describe('Array', function(){
    describe('#indexOf()', function(){
        it('should return -1 when the value is not present', function(){
            assert.equal(-1, [1,2,3].indexOf(5));
            assert.equal(-1, [1,2,3].indexOf(0));
        })
    })
})


describe('TestService', function() {
    var test, httpBackend;

    beforeEach(module('gogrinder'));

    beforeEach(inject(function(TestService, $httpBackend) {
        test = TestService;
        httpBackend = $httpBackend;
    }));

    describe('exit()', function() {
        it('should call DELETE on REST service method /stop', function() {
             httpBackend
                .expect('DELETE', 'http://localhost:3000/stop')
                .respond(200, '');

            test.exit();

            httpBackend.flush();
        });
    });
});
