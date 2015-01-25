package main

const index = `

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="">
    <meta name="author" content="">
    <link rel="icon" href="../../favicon.ico">


    <title>Dashboard Template for Bootstrap</title>

    <!-- Bootstrap core CSS -->
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css" rel="stylesheet">

    <!-- Custom styles for this template -->
    <link href="dashboard.css" rel="stylesheet">

    <!-- Just for debugging purposes. Don't actually copy these 2 lines! -->
    <!--[if lt IE 9]><script src="../../assets/js/ie8-responsive-file-warning.js"></script><![endif]-->
    <script src="../../assets/js/ie-emulation-modes-warning.js"></script>

    <!-- HTML5 shim and Respond.js for IE8 support of HTML5 elements and media queries -->
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->

    <style media="screen" type="text/css">

      .label-INACTIVE {
        background-color: #5bc0de;
      }
      .label-true {
        background-color: #337ab7;
      }
      .label-ACTIVE {
        background-color: #d9534f;
      }
      .label-INSTALLED {
        background-color: #5cb85c;
      }
    </style>
  </head>

  <body>

          <h2 class="sub-header">Status:Hosts</h2>
          <div class="table-responsive">
            <table class="table table-condensed">
              <thead>
                <tr>
                  <th>Hostname</th>
                  <th>Preinstall</th>
		              <th>Install</th>
                  <th>Status</th>
                  <th>Macaddress</th>
                </tr>
              </thead>
              <tbody>
              {{range.}}
                <tr class="{{if .Active}}active{{end}}">
                  <td>{{.Name}}</td>
                  <td>{{.Preinstall}} 
                    <div class="btn-group btn-group-xs" role="group" aria-label="...">
			                 <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Macaddress}}/preinstall/raw"><span class="glyphicon glyphicon-search" aria-hidden="true"></span></a>
                       <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Macaddress}}/preinstall/preview"><span class="glyphicon glyphicon-cloud-download" aria-hidden="true"></span></a>
                    </div> 
		               </td>
                  <td>{{.Install}}  
		            	  <div class="btn-group btn-group-xs" role="group" aria-label="...">
  		        		    <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Macaddress}}/install/raw"><span class="glyphicon glyphicon-search" aria-hidden="true"></span></a>
				              <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Macaddress}}/install"><span class="glyphicon glyphicon-cloud-download" aria-hidden="true"></span></a>
			             </div>
		              </td>
		            <td>
                <span class="label label-{{.Status}}">{{.Status}}</span> 
                  <div class="btn-group btn-group-xs" role="group" aria-label="...">
                    <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Name}}/toggle"><span class="glyphicon glyphicon-retweet" aria-hidden="true"></span></a>
                  </div>
                </td>
                  <td>{{.Macaddress}} 
                  <div class="btn-group btn-group-xs" role="group" aria-label="...">
                    <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Name}}/select"><span class="glyphicon glyphicon-star" aria-hidden="true"></span></a>
                    <a type="button" class="btn btn-default btn-xs" href="/v1/{{.Macaddress}}/inspect"><span class="glyphicon glyphicon-search" aria-hidden="true"></span></a>
                  </div>
                  </td>
                </tr>
              {{end}}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Bootstrap core JavaScript
    ================================================== -->
    <!-- Placed at the end of the document so the pages load faster -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.1/jquery.min.js"></script>
    <script src="../../dist/js/bootstrap.min.js"></script>
    <script src="../../assets/js/docs.min.js"></script>
    <!-- IE10 viewport hack for Surface/desktop Windows 8 bug -->
    <script src="../../assets/js/ie10-viewport-bug-workaround.js"></script>
  </body>
</html>
`