# First steps using GO Adabas API

<!-- TOC -->

- [First steps using GO Adabas API](#first-steps-using-go-adabas-api)
  - [Prerequisite](#prerequisite)
  - [Testing](#testing)
  - [Query tool using Adabas Database id](#query-tool-using-adabas-database-id)
  - [Query tool using Adabas Map](#query-tool-using-adabas-map)

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
bin/linux_amd64/tests/testsuite -c 10000 -t 2 "167(adatcp://localhost:60001)"
```

## Query tool using Adabas Database id

A tool is provided to search in your database records using database id and two-character short names for fields. You can define your own searches and field list using the `query` test tool.

```sh
bin/linux_amd64/tests/query -d AA,F0 -f 9 -s "AA=[0:Z]" -o -l 2 "24(adatcp://adahost:60024)"
```

This query will search for all records with field `AA` starting with 0 to Z and it will read field AA and F0. The number of records is limited to two records.
The corresponding output of the `query` command will be like following:

```sh
Start thread 1/1
Result of query search= descriptor= and fields=AA,F0Dump all result values
Record Isn: 0001
  A0 = [ 1 ]
   AA = > 50005800 <
  F0 = [ 1 ]
   FA[01] = [ 1 ]
    FA[01,01] = > 26 Avenue Rhin Et Da                                         <
   FB[01] = > Joigny                                   <
   FC[01] = > 89300      <
   FD[01] = > F   <
   F1[01] = [ 1 ]
    FE[01] = > 1033   <
    FF[01] = > 44864858        <
    FG[01] = >                 <
    FH[01] = >                 <
    FI[01] = [ 0 ]
Record Isn: 0002
  A0 = [ 1 ]
   AA = > 50005600 <
  F0 = [ 1 ]
   FA[01] = [ 1 ]
    FA[01,01] = > 51 Rue Victor Faugie                                         <
   FB[01] = > Vienne                                   <
   FC[01] = > 38200      <
   FD[01] = > F   <
   F1[01] = [ 1 ]
    FE[01] = > 1033   <
    FF[01] = > 42457727        <
    FG[01] = >                 <
    FH[01] = >                 <
    FI[01] = [ 0 ]
Finish thread 1 with 1 loops
Done testsuite test took 6.913943ms
```

## Query tool using Adabas Map

A tool is provided to search in the Adabas database using new Adabas Map with long names for fields. You can define your own searches and field list using the `querym` test tool.

```sh
bin/linux_amd64/tests/querym -r 24,4 -d personnel-id,full-name -s "name=SMITH" -o -l 2 "EMPLOYEES"
```

This query will search for all records with field `name` equals to `SMITH` and it will read field personnel-id and all entries of the `full-name` group. The number of records is limited to two records.
The corresponding output of the `query` command will be like following:

```sh
Start thread 1/1
Search for  name='SMITH'
Result of query search=name='SMITH' descriptor= and fields=personnel-id,full-name
Dump all result values
Record Isn: 0579
  personnel-data = [ 1 ]
   personnel-id = > 20009300 <
  full-name = [ 1 ]
   first-name = > Seymour <
   middle-name = > C. <
   name = > SMITH <
Record Isn: 0634
  personnel-data = [ 1 ]
   personnel-id = > 20015400 <
  full-name = [ 1 ]
   first-name = > Ann <
   middle-name = > Phyllis <
   name = > SMITH <
Finish thread 1 with 1 loops
Done testsuite test took 557.644315ms
```