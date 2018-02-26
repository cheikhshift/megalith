function SavvedSetting(data) {
    if (data.result) {
        Snackbar("Settings saved!");
    }
}


app.controller('Settings', ['$scope', function($scope) {

    mdc.tabs.MDCTabBarScroller.attachTo(document.querySelector('#my-mdc-tab-bar-scroller'));
    $scope.mega = {};
    $scope.loading = false;
    $scope.selected = 0;

    Mega(MegaCB.bind($scope));

    $scope.updateMail = () => {
        UMail($scope.mega.Mail, SavvedSetting.bind($scope))
    }

    $scope.updateTw = () => {
        UTw($scope.mega.SMS, SavvedSetting.bind($scope))
    }
}]);