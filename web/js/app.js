let drawer;

function get(name) {
    if (name = (new RegExp('[?&]' + encodeURIComponent(name) + '=([^&]*)')).exec(location.search))
        return decodeURIComponent(name[1]);
}

window.Snackbar = (message, action, handler) => {
    //window.__snackbar.setAriaHidden();
    const dataObj = {
        message: message,
        actionText: action,
        actionHandler: handler,
        timeout: 2500
    }
    window.__snackbar.show(dataObj);
}

$(document).ready(function() {
    drawer = new mdc.drawer.MDCPersistentDrawer(document.querySelector('.mdc-drawer--persistent'));
    document.querySelector('.menu').addEventListener('click', () => drawer.open = drawer.open ? false : true);

});

function MegaCB(data) {
    if (data.result) {
        this.mega = data.result;
       
        var edit = get("edit");
        if (edit && this.editServer) {
            for (var i = this.mega.Servers.length - 1; i >= 0; i--) {
             
                if (this.mega.Servers[i].ID == edit) this.editServerInt(i)
            }
        }
        if (this.aggData){
          this.aggData();
        }
        this.$apply();
    } else {
        disconn();
    }
}

function disconn() {
    Snackbar("Megalith is down!!!");
}

function rndFunc(val) {
  return  Math.round(val * 100) 
}


//Start of angular controller
var app = angular.module('app', []);

app.controller('dashboard', ['$scope', function($scope) {
    $scope.mega = {};
    $scope.loading = false;
    $scope.round = rndFunc;
    $scope.aggData = () => {
        if (!$scope.mega.Servers) return;
        for (var i = $scope.mega.Servers.length - 1; i >= 0; i--) {
          var server = $scope.mega.Servers[i];
          setTimeout(function(server){
            GetLog(server,function(res){
              if(res.result){
                var ctx = document.getElementById(server.ID).getContext('2d');
                  data = {
                    datasets: [{
                        data: [],
                        backgroundColor: ["#333", "#0074D9", "#FF4136", "#2ECC40", "#FF851B", "#7FDBFF", "#B10DC9", "#FFDC00", "#001f3f", "#39CCCC", "#01FF70", "#85144b", "#F012BE", "#3D9970", "#111111", "#AAAAAA"]
                    }],
                    labels: [ ]
                };
                for (var o = res.result.Requests.length - 1; o >= 0; o--) {
                   var req = res.result.Requests[o];
                   var indexofcode = data.labels.indexOf(`Code: ${req.Code}`);
                   if (indexofcode == -1){
                      data.labels.push(`Code: ${req.Code}`)
                      data.datasets[0].data.push(1)
                   } else {
                      data.datasets[0].data[indexofcode]++;
                   }
                }

                var myDoughnutChart = new Chart(ctx, {
                      type: 'doughnut',
                      data: data,
                      options: {}
                  });
              }
            }) 
        }, (200 * i),server);
        }
       
    }

    Mega(MegaCB.bind($scope));


}])





//momentum functions

function jsrequestmomentum(url, payload, type, callback) {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = () => {
        if (xhttp.readyState == 4) {
            var success = (xhttp.status == 200)
            if (type == "POSTJSON") {
                try {
                    callback(JSON.parse(xhttp.responseText), success);
                } catch (e) {
                  console.log(e)
                    console.log("Invalid JSON");
                    callback({ error: xhttp.responseText == "" ? "Server wrote no response" : xhttp.responseText }, false)
                }
            } else callback(xhttp.responseText, success);
        }
    };

    var serialize = (obj) => {
        var str = [];
        for (var p in obj)
            if (obj.hasOwnProperty(p)) {
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
            }
        return str.join("&");
    }
    xhttp.open(type, url, true);

    if (type == "POST") {
        xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        xhttp.send(serialize(payload));
    } else if (type == "POSTJSON") {
        xhttp.setRequestHeader("Content-type", "application/json");
        xhttp.send(JSON.stringify(payload));
    } else xhttp.send();
}
