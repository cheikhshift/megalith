app.controller('k8s_pods', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.spod = false;
    $scope.pods = [];
    $scope.groupMap = {};
    $scope.podConfig = false;
    $scope.usageMap = [];
    Mega(MegaCB.bind($scope));

    GetPods(function(data) {

        if (data.result)
            for (var i = data.result.length - 1; i >= 0; i--) {

                if ($scope.mega.KubeSettings.MetricNamespace == "" || $scope.mega.KubeSettings.MetricNamespace == data.result[i].metadata.namespace)
                    for (var o = data.result[i].containers.length - 1; o >= 0; o--) {
                        var container = data.result[i].containers[o]
                        var pod = { Name: container.name }

                        if (!$scope.groupMap[container.name]) {
                            $scope.groupMap[container.name] = 1
                            $scope.pods.push(pod)
                        } else $scope.groupMap[container.name]++

                            $scope.usageMap.push(container)
                    }
            }
        $scope.$apply();
    })


    $scope.SetConfig = (name) => {
        if ($scope.mega.KubeSettings.Monitoring) {
            for (var i = $scope.mega.KubeSettings.Monitoring.length - 1; i >= 0; i--) {
                if ($scope.mega.KubeSettings.Monitoring[i].Name == name) {
                    $scope.podConfig = $scope.mega.KubeSettings.Monitoring[i]
                    return
                }

            }
            $scope.addpod(name)
        } else {
            $scope.addpod(name)
        }
    }

    $scope.addpod = (name) => {
        AddPod({ Name: name }, function(data) {
            $scope.mega.KubeSettings.Monitoring = data.watching
            $scope.SetConfig(name)
            $scope.$apply()
        })
    }

    $scope.editPod = function(pod) {
        $scope.spod = pod;
        $scope.SetConfig(pod.Name)
    }

    $scope.updatePod = function(pod) {
        UpdatePod(pod, function(data) {
            if (data.result) {
                Snackbar("Pod saved!");
            }
        })
    }


}])