# Exploit your assets in Adabas by using the Go Adabas-API

<!-- TOC -->

- [Exploit your assets in Adabas by using the Go Adabas-API](#exploit-your-assets-in-adabas-by-using-the-go-adabas-api)
  - [Introduction](#introduction)
  - [Usage](#usage)
  - [Adabas Go API example](#adabas-go-api-example)
  - [New Map repository](#new-map-repository)
  - [First step](#first-step)
  - [Summary](#summary)

<!-- /TOC -->

## Introduction

This is the Adabas API for the programming language Go. It supports data access to the Software AG Adabas database. A detailed overview about the design and technical implementation you can find [here](.//doc//Overview.md)

## Usage

Adabas Go API can be download using the `go get` command:

```bash
go get -u github.com/softwareag/adabas-go/adabas
```

You can compile it with the Adabas TCP/IP interface only or using Adabas local access with Adabas Client native libraries. By default the Adabas TCP/IP interface is compiled only. To enable Adabas Client native link to Adabas you need to provide the Go build tag `adalnk` and the CGO compile flags defining build flags for the Adabas Client library. If the Adabas environment is sourced, you can define CGO compile flags as following:

```sh
CGO_CFLAGS = -I$(ACLDIR)/inc
CGO_LDFLAGS = -L$(ACLDIR)/lib -ladalnkx -lsagsmp2 -lsagxts3 -ladazbuf
```

## Adabas Go API example

A quick example to read data from a database file 11 of Adabas database with database id 23 is like here

```go
connection, err := NewConnection("acj;target=23")
if err!=nil {
  return
}
defer connection.Close()
connection.Open()
readRequest, rErr := connection.CreateReadRequest(11)
readRequest.QueryFields("AA,AB")
readRequest.Limit = 0
result,err := request.ReadLogicalWith("AA=60010001")
```

See detail documentation [here](.//doc//README.md)

## New Map repository

The logical view to the data can be defined using Adabas maps. A detailed description about Adabas maps is described [here](.//doc//AdabasMap.md)

The infrastructure of the Java API for Adabas (Adabas Client for Java) can be used. To create logical views and field references, the Java API provides Adabas maps. The Adabas Data Designer rich client or Eclipse plugin can be used to define Adabas map definitions. A programmical approach to create a Adabas Map is part of the Adabas Go API.

This is an example using a logical view to read the data out of Adabas. In the next example the Adabas read operation is using Adabas maps

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

See detail documentation [here](.//doc//AdabasMap.md)

## First step

A detailed description how to do the first steps using the Adabas Docker community edition is provided [here](.//doc//FirstSteps.md).
Independent of the used environment of Docker (like Kubernetes or others), it describe how to call Adabas.

## Summary

The Go Adabas-API offers easy access to store or read data in or out of Adabas. In advance the transactionality is provided. The Go API should help developers to work with data in Adabas without having the need of being an Adabas expert knowing special Adabas Database features.
Go functions enable developers to use Go as a programming language accessing Adabas in the same way other data sources are embedded in a Go project.

______________________
These tools are provided as-is and without warranty or support. They do not constitute part of the Software AG product suite. Users are free to use, fork and modify them, subject to the license agreement. While Software AG welcomes contributions, we cannot guarantee to include every contribution in the master project.	

Contact us at [TECHcommunity](mailto:technologycommunity@softwareag.com?subject=Github/SoftwareAG) if you have any questions.

