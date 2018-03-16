app.controller('k8s_setup', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.scontact = false;

    Mega(MegaCB.bind($scope));


    $scope.saveConfig = function() {
        UpdateKubernetes($scope.mega.KubeSettings, function(data) {
            if (data.result) {
                Snackbar("Configuration saved!");
            }
        })
    }



}])