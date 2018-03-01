app.controller('Contacts', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.scontact = false;

    Mega(MegaCB.bind($scope));

    $scope.addContact = function() {
        AddContact(function(data) {
            $scope.mega.Contacts = data.result
            Snackbar("+ New Contact ")
            $scope.$apply();
        })
    }



    $scope.editContact = function(Contact) {
        $scope.scontact = Contact;
        $scope.sendpoint = false;
    }

    $scope.deleteContact = function(Contact) {
        DContact(Contact, function(data) {

            if (data.result) {
                $scope.scontact = false;
                Snackbar("Contact removed ");
                $scope.mega.Contacts = data.result;
                $scope.$apply();
            }
        })
    }

    $scope.updateContact = function(Contact) {
        UContact(Contact, function(data) {
            if (data.result) {
                Snackbar("Contact saved!");
            }
        })
    }



}])