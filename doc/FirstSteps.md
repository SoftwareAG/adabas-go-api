# First steps using GO Adabas API

<!-- TOC -->

- [First steps using GO Adabas API](#first-steps-using-go-adabas-api)
	- [Prerequisite](#prerequisite)
	- [Testing](#testing)

<!-- /TOC -->

## Prerequisite

A basic prerequisite is the Adabas database. The GO Adabas API can use the native Adabas client library to access the Adabas database using local and Entire Network as a destination. Corresponding configuration of the Adabas Client is needed.

Beside the native access to Adabas, the GO Adabas API supports the new Adabas TCP/IP link introduces in Adabas v6.7. With the Adabas TCP/IP a direct connection to the Adabas nucleus can be established.

To test the Adabas you can use the Adabas Community Edition for Docker. The container are available in the Docker store. Only registration is required.
Since Adabas version v6.7 the Docker image contains the Adabas TCP/IP link. You can start the Docker container like this

```sh
docker run -d -p 60001:60001 -p 8191:8191 -e ACCEPT_EULA=Y -e ADADBID=12 -e ADA_DB_CREATION=demodb --name adabas-db store/softwareag/adabas-ce:6.7.0
```

This Docker command will start an Adabas database nucleus in an Docker container. The Adabas database will contain Adabas demo data. The Adabas TCP/IP link will listen on port 60001. In adavance the Adabas RESTful administration can be exposed. In the example above the RESTful service is exposed on port 8191.

For a detailed list of environments variables of the Adabas Docker container please see the installation instructions in the Docker store.

## Testing

Part of the GO code is a program called `testsuite`. This program can access the Adabas example file 11, the EMPLOYEES example. This command will start the EMPLOYEES search for all records containing the name `SMITH` (`NAME=SMITH`). The request to the database can be request using the  Adabas TCP/IP at the Docker container host at the Adabas TCP/IP listening port (in the example above it is 60001). This example search will call the search 10000 times in two separated threads:

```sh
bin/tests/testsuite -c 10000 -t 2 "167(adatcp://localhost:60001)"
```
