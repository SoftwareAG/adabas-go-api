# Design overview

<!-- TOC -->

- [Design overview](#design-overview)
    - [Background and Benefits](#background-and-benefits)
    - [General Considerations](#general-considerations)
    - [Feature set](#feature-set)
    - [Technical Overview](#technical-overview)
    - [Object of migration](#object-of-migration)

<!-- /TOC -->

## Background and Benefits

Go is the de facto standard in terms of programming in Microservice environments. Besides many other languages like Java, C\#, Perl, Python, Ruby, Go is getting popular
(see <https://www.tiobe.com/tiobe-index/>).

To be part of the Go community all data stores are forced to provide a Go-based interface to access data.
The Adabas Go API provides a possibility to maintain and handle Adabas Metadata.
Beside Java, Natural and 3GL languages it is compulsory and a logical way to have a state-of-the-art Go based interface for Adabas to be considered as a data store that can be easily used in projects.

The new interface addresses developers and software designer who work primarily with Go and related concepts. However, developers are used to work with other programming languages can benefit from this approach. The usage is similar to the Adabas Client for Java product Software AG provides to work in Java environments.

The Go Adabas-API provides

- Make it easy for everyone to write Go-based application and interact with an Adabas database.
- Use Go methods for data access.

## General Considerations

Before diving deeper into technology let's explore what are the key points and where are the benefits. The intension is to offer an easy-to-use Adabas interface covering all functions of the native Adabas call interface, which is very powerful but hard to code. Of course, it is possible to use the Adabas call interface based on different raw buffer types in a Go program. A much more sophisticated way is to take the higher-level Adabas-API using metadata to access the Adabas database.

Another important point is usability in general, but especially for developers who are not familiar with Adabas. Using the interface makes it easy to write applications based on data stored in Adabas. This is not only a topic for Linux, Unix, Windows (LUW) based organization but also for companies running on mainframe e.g. z/OS.

Performance is always a topic. The interface is designed to provide optimal throughput and performance of the resources while processing Adabas data and communicating with Adabas. A positive side-affect is that the interface allows an easy integration of Adabas into other new technology like e.g. Microservice.

Actually not compulsory for an interface, but worth to mention, the interface fully supports Adabas transactions and the entire set of access methods and currently almost all of the special descriptors of Adabas.

## Feature set

Software AG does not guarantee support if there are issues with the Go Adabas API. 
All current Adabas database versions are accessable. On Adabas Mainframe there might be some issues. Current features include

- Read, insert, update or delete records. Reference data in Adabas fields with Adabas short name or logical field names defined in Adabas maps
  - Logical read using Adabas search
  - Read in Adabas file physical order
  - Read in Adabas file using the ISN order
- Read and Update "period groups" and "multiple fields". You can read and update the complete record not one specific entry index
- Adabas large objects read but without partial reads
- Support to access database
  - Using the native local Adabas client connection
  - Implements the new  Adabas TCP/IP link to the database
- Provides to set various Go log frameworks to trace operations

## Technical Overview

The graphic below shows how the Go Adabas-API works in general.

![Technical Overview](.//media/Go-Design.png)

The client interface provides a set of functions which can be used in Go programs. A detailed description is avaiable [here](.//README.md). Software components are available to parse the queries and build internal calls based on the Adabas native interface. The transformed requests are sent to Adabas using the normal low level interface, the Adabas link module. This in turn allows including Entire Net-Work as middleware to reach remote databases. Ideally, the software runs local to the database. However, to reach a mainframe database Entire Net-Work is required.

The test package which is part of the sources provides a selection of example Go programs that show the usage. All tests are based on the Adabas demo file Employees.

## Object of migration

Most customers using Adabas have worked with Natural in the past. A important matter of migration are the application using Natural or Predict logical views.
Adabas provides to use Adabas short name field definitions and database id. To provide a logical name of a database file and a logical name reference to the Adabas short name, a new Adabas Map is introduced. This Adabas Map will be stored inside the Adabas database in an extra Adabas repository file.
Tools managing the Adabas maps are part of the Adabas Client for Java product Software AG provides.

The new Adabas Map technology is needed because the Natural and Predict storage of logical views are outside the Adabas database. To use the logical view independent to Natural, a storage in Adabas is elementary. The benefit is that the logical view and the data can now be part of a backup strategy and multiple logical views, dependent on the work case, can be provided.

Following import file formats are supported

- Natural SYSTRANS files including DDM definitions exported by Natural
- Remarks in the file definition files (FDT) can be used to define long names
- XML format defining the logical view