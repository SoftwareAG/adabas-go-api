# Exploit your assets in Adabas by using the Adabas API for Go

<!-- TOC -->

- [Exploit your assets in Adabas by using the Adabas API for Go](#exploit-your-assets-in-adabas-by-using-the-adabas-api-for-go)
  - [Introduction](#introduction)
  - [Features](#features)
  - [Usage](#usage)
    - [Enable native library access](#enable-native-library-access)
  - [First step](#first-step)
  - [Adabas API for Go example](#adabas-api-for-go-example)
    - [Standard usage](#standard-usage)
    - [Classic database usage](#classic-database-usage)
    - [Using a Go struct](#using-a-go-struct)
  - [Log output](#log-output)
  - [Summary](#summary)

<!-- /TOC -->

## Introduction

This package is designed for using Adabas databases from Go. It's an data interface to Adabas. You can find a detailed overview about the design and technical implementation [here](.//doc//Overview.md).

For details have a look at the API documentation. It can be referenced here: <https://godoc.org/github.com/SoftwareAG/adabas-go-api/adabas>

## Features

In general, users of the Adabas API for Go do **not** need to know a basic level of Adabas APIs such as Adabas control blocks or Adabas buffer layouts.
The Adabas API for Go provides a more user friendly and optimized use of Adabas files and fields.
It is similar to the `Adabas Client for Java` product delivered by Software AG.

This is a list of features which the Adabas API for Go supports:

- Read,  Insert, Delete and Update of Adabas records
- [Search](.//doc//QUERY.md) are limited to the Adabas standard search capabilities (see [Search](.//doc//QUERY.md) documentation)
- Work with field descriptors and special descriptors like sub-/super-descriptors
- Work with the Adabas TCP/IP (ADATCP) layer on Linux, Unix and Windows
- Work with interface to Adabas Client library (ADALNK)
- Work with Adabas Mainframe using the Entire Network infrastructure (ADALNK)
- Support Unicode access of Adabas Unicode fields
- Work with Adabas Maps, a short name to long name definition
- Work with Adabas period groups and multiple fields
- Work with large object reads and writes
- Work with partial large object reads and writes
- Is optimized for network access using ADATCP by using Multifetch feature
- Provide Go structure usage by reflecting structure fields to Adabas Map definitions

## Usage

Inside the code the Adabas API for Go can be used importing the Go API. Beside the API, some small sample applications are provided in GitHub. These examples of the Adabas API for Go can be downloaded using the `go get` command:

```bash
go get -u github.com/softwareag/adabas-go-api/adabas
```

You can compile it with the new Adabas TCP/IP interface on Linux, Unix and Windows. In this case no additional native library is needed. This is the default behavior.

Alternatively the Adabas native local access with Adabas client native libraries (ADALNK) can be used. The `AdabasClient` installation is a prerequisite in this case.

### Enable native library access

The default is the Adabas TCP/IP interface enabled in the Go API. It is possible to enable Adabas Client native interface support using the Adabas Client library. You need to provide the Go build tag `adalnk` and the CGO compile flags defining build flags for the Adabas Client library. If the Adabas environment is sourced, you can define CGO compile flags as follows:

On Unix:

```sh
CGO_CFLAGS=-I${ACLDIR}/inc
CGO_LDFLAGS=-L${ACLDIR}/lib -ladalnkx -lsagsmp2 -lsagxts3 -ladazbuf
```

On Windows:

```bat
CGO_CFLAGS=-I%ACLDIR%\..\inc
CGO_LDFLAGS=-L%ACLDIR%\..\bin -L%ACLDIR%\..\lib -ladalnkx
```

The application is built with Adabas API for Go as follows (please note that the tag `adalnk` is needed to enable local IPC access):

```go
go build -tags adalnk application.go
```

## First step

A detailed description how to do the first steps using the Adabas Docker community edition is provided [here](.//doc//FirstSteps.md).
Independent of the used environment of Docker (like Kubernetes or others), it describes how to call Adabas in a Microservice environment.

## Adabas API for Go example

### Standard usage

The logical view of the data can be defined using Adabas Maps. The Adabas Maps are introduced in the Adabas Client for Java product delivered by Software AG. A detailed description of Adabas maps is available [here](.//doc//AdabasMap.md).

The creation of Adabas maps is done by the infrastructure of the Adabas Client for Java. The product contains a Adabas Data Designer rich client or Eclipse plugin. The tools provide the management of Adabas map definitions. A programmatical approach to create Adabas maps is part of the Adabas API for Go.

In the first example a logical read on the database file uses Adabas maps:

```go
import (
  "github.com/SoftwareAG/adabas-go-api/adabas"
)
// Create new connection handler using the Adabas Map repository in database 24 file 4
connection, cerr := adabas.NewConnection("acj;map;config=[24,4]")
if cerr != nil {
  return
}
defer connection.Close()
// Create a read request using the Map definition
request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
// Define the result records content
request.QueryFields("NAME,PERSONNEL-ID")
request.Limit = 2
// Read logical using a range search query
result,rerr := request.ReadLogicalWith("PERSONNEL-ID=[11100301:11100303]")
// Result is dumped to stdout
result.DumpValues()
```

The next example shows a cursor read  on the database file:

```go
// Read logical with cursor using a range search query
result,rerr := request.ReadLogicalWithCursoring("PERSONNEL-ID=[11100301:11100303]")
// Result is dumped to stdout
for col.HasNextRecord() {
  record, rerr := col.NextRecord()
    ...
}
```

See detailed documentation [here](.//doc//AdabasMap.md).

### Classic database usage

It is possible to use classical Adabas database parameters to read data. The approach does not contain format enhancements like Natural DDM's or the Adabas Map approach. It is a raw access to the database.

A quick example to read data from a database file 11 of the Adabas database with database ID 23 is here:

```go
// Create new connection handler to database
connection, err := adabas.NewConnection("acj;target=23")
if err!=nil {
  return
}
defer connection.Close()
connection.Open()
// To work on file 11 create corresponding read request
request, rErr := connection.CreateFileReadRequest(11)
// Define the result records content
request.QueryFields("AA,AB")
request.Limit = 0
// Read in the database using search query
result,err := request.ReadLogicalWith("AA=60010001")
var aa,ac,ad,ae string
// Read given AA(alpha) and all entries of group AB to string variables
result.Values[0].Scan(&aa,&ac,&ad,&ae)
```

The example code is referenced [here](.//tests//simple_read.go). See detailed documentation [here](.//doc//README.md).

### Using a Go struct

The Adabas API for Go can handle simple Go struct definitions to map them to a Adabas Map definition.

For example if the structure is defined like this:

```go
type Employees struct {
ID        string
Birth     int64
Name      string `adabas:"Name"`
FirstName string `adabas:"FirstName"`
}
```

The struct can be used to read or store data of the field `Name` and `FirstName` directly. The fields `Birth` is wrote too.

To store the struct record, do this:

```go
storeRequest, err := adabas.NewStoreRequest(Employees{}, ada, repository)
e:=  &Employees{ID: "ID3", Birth: 456, Name: "Name3", FirstName: "First name3"}
err = storeRequest.StoreData(e)
err = storeRequest.EndTransaction()
```

The read of the struct data will be done with:

```go
request, err := adabas.NewReadRequest(Employees{}, adabas, mapRepository)
defer request.Close()
result, err := request.ReadLogicalWith("ID>'ID'")
e := result.Data[0].(*Employees)
```

All fields of the struct are mapped to an Adabas Map field name. The `adabas` tag of the struct definition changes the mapped name.

## Log output

To enable log output in sample executables, please set the `ENABLE_DEBUG` environment variable to 1 for `debug` level output and 2 for `info` level output. This will enable the logging.

To use logging in your code with the Adabas API, you can enable logging by setting the log instance with your `logger` instances with:

```go
adatypes.Central.Log = logger
```

## Summary

The Adabas API for Go offers easy access to store or read data in or out of Adabas. The Go API should help developers to work with data in Adabas without needing to be an expert on special Adabas database features.
Go functions enable developers to use Go as a programming language to access Adabas in the same way as other data sources are embedded in a Go project.
By using the native `AdabasClient` library, you can access all platforms Adabas runs on like Linux, Unix, Windows and Mainframe (z/OS with Entire Network).
Step by step all relevant Adabas features are supported.

______________________
These tools are provided as-is and without warranty or support. They do not constitute part of the Software AG product suite. Users are free to use, fork and modify them, subject to the license agreement. While Software AG welcomes contributions, we cannot guarantee to include every contribution in the master project.
______________
For more information you can Ask a Question in the [TECHcommunity Forums](https://tech.forums.softwareag.com/tag/adabas).

You can find additional information in the [Software AG TECHcommunity](http://techcommunity.softwareag.com/home/-/product/name/adabas).
______________
Contact us at [TECHcommunity](mailto:technologycommunity@softwareag.com?subject=Github/SoftwareAG) if you have any questions.
