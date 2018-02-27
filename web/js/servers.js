app.controller('servers', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.sserver = false;
    $scope.sendpoint = false;
    $scope.round = rndFunc;
   

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
        $scope.sserver.Endpoints.push({ Method: 'GET' });
        Snackbar("+ Endpoint");
    }

    $scope.editEndpoint = function(endpoint) {
        $scope.sendpoint = endpoint;
    }

    $scope.removeEndpoint = function(endpoint) {
        $scope.sserver.Endpoints.splice($scope.sserver.Endpoints.indexOf(endpoint), 1);
        $scope.sendpoint = false;
        Snackbar("Endpoint removed!");
    }

     Mega(MegaCB.bind($scope));

}])