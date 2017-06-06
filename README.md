# USAGE-API
The projects aims to display the consumption data for different users. The data is a list of electricity consumption and temperature values per days and months. Each user has their own data, of course, and no user shall be able to see data belonging to other users.

## ENDPOINTS
The project exposes below endpoints for the user to view there electricity consumption data. All the requests are made over `https` and application expects a `Authorization Header` to be set. The application uses `Basic Authorization`



1. **/ping**: It is used to check the health of the application.

2. **/limits** : This endpoint is used to fetch the maximum and minimum values for the various attributes of the data like `temperature`,`consumption` etc.

3. **/data** : This endpoint accepts various query params to provide data over a time range for the user.

**Example** : curl -XGET --user username1:password1 https://localhost:8080/data?resolution=M&count=3&start=2014-02-03 --cacert ./cert/cert.pem

In order to communicate over https we need to pass location of the CA certificate generated for this application.



## RUN
In order to run the project, we need `golang` installed. The project is tested against `go v1.8`. Once go is installed and `GOPATH` is set correct below steps are needed to be followed to run the project.

1. `go get` to install the dependencies.
2. `go run main.go` which will run the main file and start the server.


## TEST
Tests are provided to test the api. In order to run the tests  we need to invoke `go test -v` which will execute the tests in verbose mode. Each of the tests are self contained and add and clean up the data once the test case is executed.


## NOTE
In order to keep things simple, there are only two users added in the production database. 
