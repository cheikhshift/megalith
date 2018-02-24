function SavvedSetting(data){
	if (data.result) {
                Snackbar("Settings saved!");
    }
}

app.controller('Settings', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;

    Mega(MegaCB.bind($scope));

    $scope.updateMail = () => {
    	UMail($scope.mega.Mail, SavvedSetting.bind($scope))
    }

    $scope.updateMega = () => {
    	UInt($scope.mega.Cl, SavvedSetting.bind($scope))
    }
}]);