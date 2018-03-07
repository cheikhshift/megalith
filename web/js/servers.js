app.controller('servers', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.sserver = false;
    $scope.sendpoint = false;
    $scope.round = rndFunc;
    $scope.loadingChart = false;

    //set chart context info
    var ctx = document.getElementById("chartdispArea").getContext('2d');

    $scope.addServer = function() {
        AddServer(function(data) {
            $scope.mega.Servers = data.result
            Snackbar("+ New server ")
            $scope.$apply();
        })
    }



    $scope.editServer = function(server) {
        $scope.sserver = server;
        $scope.sendpoint = false;
    }

    $scope.editServerInt = function(server) {
        $scope.sserver = $scope.mega.Servers[server];
        $scope.sendpoint = false;
    }

    $scope.deleteServer = function(server) {
        DServer(server, function(data) {
            if (data.result) {
                $scope.sserver = false;
                Snackbar("Server removed ");
                $scope.mega.Servers = data.result;
                $scope.$apply();
            }
        })
    }

    $scope.updateServer = function(server) {
        UServer(server, function(data) {
            if (data.result) {
                Snackbar("Server saved!");
            }
        })
    }

    $scope.addEndpoint = function() {
        if (!$scope.sserver.Endpoints) $scope.sserver.Endpoints = [];
        $scope.sserver.Endpoints.push({ Method: 'GET', Uptime: 0 });
        Snackbar("+ Endpoint");
    }

    $scope.editEndpoint = function(endpoint) {
        $scope.sendpoint = endpoint;
        //chartdispArea
        $scope.loadingChart = true;
        var APIID = `${endpoint.Method}${endpoint.Path}`
        GetLog($scope.sserver, function(res) {

            if (res.result) {
                data = {
                    datasets: [{
                        data: [],
                        backgroundColor: chartColors
                    }],
                    labels: []
                };
                $scope.loadingChart = false;
                $scope.$apply();
                if (!res.result.Requests) {
                    return;
                }
                for (var o = res.result.Requests.length - 1; o >= 0; o--) {
                    var req = res.result.Requests[o];
                    if (req.Owner == APIID) {
                        var indexofcode = data.labels.indexOf(`Code: ${req.Code}`);
                        if (indexofcode == -1) {
                            data.labels.push(`Code: ${req.Code}`)
                            data.datasets[0].data.push(1)
                        } else {
                            data.datasets[0].data[indexofcode]++;
                        }
                    }
                }

                var myDoughnutChart = new Chart(ctx, {
                    type: 'doughnut',
                    data: data,
                    options: {}
                });

            }
        })
    }

    $scope.removeEndpoint = function(endpoint) {
        $scope.sserver.Endpoints.splice($scope.sserver.Endpoints.indexOf(endpoint), 1);
        $scope.sendpoint = false;
        Snackbar("Endpoint removed!");
    }

    Mega(MegaCB.bind($scope));


}])