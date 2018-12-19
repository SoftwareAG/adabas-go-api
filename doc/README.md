# Technical overview about Go API

<!-- TOC -->

- [Technical overview about Go API](#technical-overview-about-go-api)
  - [Content](#content)
  - [Build](#build)
  - [Usage](#usage)
  - [Database connection string](#database-connection-string)
  - [Request types](#request-types)
  - [Example usage](#example-usage)
    - [Search Adabas](#search-adabas)
    - [Modification request](#modification-request)
    - [Search syntax](#search-syntax)
  - [Dynamic result function](#dynamic-result-function)
  - [Transactions](#transactions)

<!-- /TOC -->

## Content

This is a documentation describing the Go based Adabas-API. The API can
provide access to Adabas database data. It is the Go-based implementation of the Adabas Client for Java delivered by Software AG.

This Go implementation contains various access methods to query Adabas database content.

This implementation can access Adabas in various ways

- The local database IPC communication using the native Adabas client library
- The new Adabas TCP/IP access without using the Adabas client

The remote Adabas TCP/IP connection does not use the Adabas client library. To access local Adabas databases, the native Adabas Client library is prerequisite to compile with.

## Build

Adabas Go API can be download using the `go get` command:

```bash
go get -u github.com/softwareag/adabas-go/adabas
```

You can compile it with the Adabas TCP/IP interface only or using Adabas local access with Adabas Client native libraries. By default the Adabas TCP/IP interface is compiled only. To enable Adabas Client native link to Adabas you need to provide the Go build tag `adalnk` and the CGO compile flags defining build flags for the Adabas Client library. If the Adabas environment is sourced, you can define CGO compile flags as following:

```sh
CGO_CFLAGS = -I$(ACLDIR)/inc
CGO_LDFLAGS = -L$(ACLDIR)/lib -ladalnkx -lsagsmp2 -lsagxts3 -ladazbuf
```

## Usage

Like the Adabas Client for Java implementation the Go Adabas API can be used with referencing

- Adabas two-character short names and database id. This is the classic referenced used by Adabas
- Using the Map definition used by the Adabas Client for Java version. Here the database name and the long name definitions can be defined by importing or defining Long names.

## Database connection string

If classic Adabas database id's reference is used, the new Adabas TCP/IP uses a database connection string to reference the remote destination database. Example of database references are

| URL  | Destination location  |
|---|---|
| "23" |  Local database 23 |
| "23(tcpip://localhost:0)" |  Local database 23, port 0 indicate local usage |
| "23(adatcp://hutzle:60001)" |  Adabas TCP/IP based connection to remote database 23 on host `hutzle` at port 60001 |

## Request types

Adabas internally works with a strict set of Adabas fields to be used in one Adabas access call. This enhance caching facilities  restrict the dynamic usage of different request field definitions. Similar to Adabas Client for Java, each request with different field sets need to have the own Request instance.

Three Request types are provided

- `ReadRequest`: needed to read or search Adabas data
- `StoreRequest`: needed to update or insert Adabas data
- `DeleteRequest`: needed to delete Adabas data

The request can be created and combined using a `Connection` instance. All these Request can be combined to an transaction. Final update, insert or delete can be finished using the end transaction call.

## Example usage

There is a set of tests call Adabas calls to the database. The database data are the demo daatabase files delivered with Adabas.

All the examples are done with the single Adabas example EMPLOYEES database file 11. The demo database contains the file. The Map definitions can be created using the Adabas Data Designer or the Mapping tool provided with  Adabas Client for Java. Only basic Map creation functionality is included.

### Search Adabas

Similar to Natural logic it is possible to request descriptor and various search variantes.

You can read data using

- logical searches based one descriptor order queries
- physical order reads based on the order in the Adabas container
- descriptor reads only accessing the index (Adabas ASSO)

Detailed documentation about search facilities are documented [here](.//QUERY.md).

Below is an example showing a logical read on an descriptor value. This is an classical Adabas access example:

```go
connection, err := NewConnection("acj;target=23(adatcp://remote:60001)")
if err!=nil {
  return
}
defer connection.Close()
connection.Open()
readRequest, rErr := connection.CreateReadRequest(11)
readRequest.QueryFields("AA,AB")
readRequest.Limit = 0
result := &RequestResult{}
err := request.ReadLogicalWith("AA=60010001", nil, result)
```

Using the Adabas Map EMPLOYEES for this query it would look like the next example. Here the Range including values for the PERSONNEL-ID from 11100301 to 11100303 including both values are searched for:

```go
connection, cerr := NewConnection("acj;map;config=[24,4]")
if cerr != nil {
  return
}
defer connection.Close()
request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
if err != nil {
  return
}
request.QueryFields("NAME,PERSONNEL-ID")
request.Limit = 2
result := &RequestResult{}
err = request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]", nil, result)
if err != nil {
  return
}
result.DumpValues()
```

The result of the last example would look like this

```output
Dump all result values
Record Isn: 0251
PERSONNEL-ID = > 11100301 <
  FULL-NAME = [ 1 ]
    NAME = > BERGMANN             <
Record Isn: 0383
PERSONNEL-ID = > 11100302 <
  FULL-NAME = [ 1 ]
     NAME = > HAIBACH              <
```

### Modification request

Similar to the query request you can update, insert and delete Adabas data.

```go
// Connection handle. Connection can contain transaction elements
connection, err := NewConnection("acj;map")
if !assert.NoError(t, err) {
  return
}
defer connection.Close()
connection.Open()
// Create a store request
storeRequest, rErr := connection.CreateMapStoreRequest("EMPLOYEES")
if rErr != nil {
   return
}
// define the record field part of the update
storeRequest.StoreFields("PERSONNEL-ID,FULL-NAME")
record, err := storeRequest.CreateRecord()
// Fill the record data
err = record.SetValueWithIndex("PERSONNEL-ID", nil, "777777")
err = record.SetValueWithIndex("FIRST-NAME", nil, "WABER")
err = record.SetValueWithIndex("MIDDLE-NAME", nil, "EMIL")
err = record.SetValueWithIndex("NAME", nil, "MERK")
err = storeRequest.Store(record)
if err != nil {
    return
}
// End of transaction, here the insert is finished
err=storeRequest.EndTransaction()
```

### Search syntax

see documentation [here](.//QUERY.md)

## Dynamic result function

Similar to the listener in Adabas Client for Java, the result data can be processed during query and does not needed to be stored in an result list lowering the memory consumpion.

By default using the ResultRecord instance, the query will be store in a list

```go
result, err := request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
```

But you can work with a function to pass structures and methods to just work the result received by the database. Here the result can be traversered field by field.

```go
func parseTestConnection(adabasRequest *parser.AdabasRequest, x interface{}) (err error) {
  parseTestStructure := x.(parseTestStructure)
  tm := parser.TraverserValuesMethods{EnterFunction: extractMapField}
  adabasRequest.Definition.TraverseValues(tm, adabasMap)
}
err = request.ReadLogicalWith("AA=[11100301:11100305]", parseTestConnection, parseTestStructure)

```

## Transactions

Inside a `Connection` instance a chain of different request can define a transaction. A read request combined with a store request can finally end using either the end of transaction or the backout transaction method.

______________________
These tools are provided as-is and without warranty or support. They do not constitute part of the Software AG product suite. Users are free to use, fork and modify them, subject to the license agreement. While Software AG welcomes contributions, we cannot guarantee to include every contribution in the master project.	

Contact us at [TECHcommunity](mailto:technologycommunity@softwareag.com?subject=Github/SoftwareAG) if you have any questions.
