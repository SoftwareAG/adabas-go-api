# Adabas Map concept

<!-- TOC -->

- [Adabas Map concept](#adabas-map-concept)
  - [Concept](#concept)
  - [Migration](#migration)
    - [Object of migration](#object-of-migration)
    - [New Map repository](#new-map-repository)
  - [Adabas Data Designer](#adabas-data-designer)
    - [Design of the Adabas Map](#design-of-the-adabas-map)
    - [Adabas Map FDT](#adabas-map-fdt)
  - [Usage of Adabas Maps](#usage-of-adabas-maps)
  - [Import and Export of maps](#import-and-export-of-maps)

<!-- /TOC -->

## Concept

"The spirits that I called". In the more then fourthy years of processing of Adabas the tool environment has evaluate different concepts. In the past metadata were expensive and so Adabas is designed to have only two-character field names. Short after that additional 4GL-based programming languages like Natural are worked out. Natural and the add-on Predict extend the two-character field names to long names.

Unfortunately customers uses the independent long name facility to provide different views of one database file. Therefore it is not very easy to integrate it into the database. In advance there were different storage places where the long name definition was rested.

The concept of Adabas Map is to consolidate all the places outside and inside the database to be in the database. This provide a possibility to integrate the metadata into an backup startegy.

But in time of continous delivery and continous testing the Adabas Map concept provides the possibilty, to define your Map configuration using an API.

## Migration

### Object of migration

Customers using Adabas work with Natural. The first important object of migration are customers using Natural or Predict. In advanced customers using Adabas without Natural are considered to use plain FDT definitions. A small possibility is provided to use Adabas short name field definitions and database id directly.

To provide a Long name database file reference and a long name reference to the Adabas short name, a new Adabas Map is introduced. In advance this Adabas Map will be stored inside the Adabas database.

### New Map repository

The Adabas Client for Java introduces a logical view of Adabas short names mapped to a long name. Various classical methods can be imported to the logical view called Adabas Maps. Using SYSTRANS Maps can be created out of

- Natural DDM
- Natural Predict defnitions
- CONNX SQL long name mapping to SQL structures

All old long to short name mapping techniques have different repositories where the mapping is defined. Except Predict, the Map definition is stored outside the database. A full backup strategy could not be provided.

To store all relevant data, metadata and bussiness data, inside the Adabas database, the corresponding new Adabas Map file is developed. A migration to the new Map for DDM's is provided. Natural Predict generates DDM. Natural provides export functionality (SYSTRANS) to migrate the DDM into the Map repository file.

A backup strategy containing Maps and Adabas data, metadata and bussiness data respectively, could be established.

## Adabas Data Designer

The Data Designer is a graphical tool to create and maintain an Adabas file that contains the Meta data and the Maps. In Adabas the Meta data are stored in so-called Field Description Tables (FDT). The Maps are based on Adabas files (FDTs) extended by Long Names. The Data Designer is part of the Adabas Client for Java installation and is not part of the Go Adabas-API.

Using the Adabas Data Designer it is possible to use available definitions. Both FDTs and Maps can be imported in different way

- import Meta data from Adabas (FDT)
- import Natural DDMs (SYSTRANS)
- import JSON based configuration exchanged and adapted through the testing and production environments (see import/export functionality below)

It is also possible to create Meta data from scratch or adapt Maps.

Adabas Files and Map of running databases are automatically shown when starting the Data Designer. Let's have a brief look into the Adabas Data Designer.

![Data Designer long name definition](.//media/image7.png)

On the left side databases, files, maps are shown in a tree view. By double-clicking an object for example "TestMapEmployee" details are shown on the right upper side. Below a data browser is included to show the data in an Adabas file. Simply mark the fields you want to see in the data browser.

### Design of the Adabas Map

The new Adabas Map design contains a enhanced format definition reclined on DDM formats.

The Adabas Map FDT is containing several metadata for the Adabas field.

| Field | Functionality | Remark |
|----|---|---|
| TA | Field indicator | 77 is the correct field indicator for Adabas Maps |
|AB|Hostname of the host the Adabas Map is created on||
|AC|Date the Adabas Map is created at|Unix timestamp|
|AD|Version of the Adabas Map|Valid version is 1|
|RN|Name of the Adabas Map||
|RF|Referenced Adabas file where the data are stored||
|RD|Referenced Adabas database|If empty data file is located on same database as the Adabas Map|
|MA|Period group containing field long name definition||
|MB|Part of MA: Short name||
|MC|Part of MA: Type of the field (extended DDM information)||
|MB|Part of MA: Short name||
|MD|Part of MA: Long name||
|ML|Part of MA: Lenght override||
|MT|Part of MA: Content type|Charset used to read Alpha fields. Needed to convert Alpha data to local charsets|
|MY|Part of MA: Format type||
|MR|Part of MA: Remarks||
|ZB|Date of modification||

### Adabas Map FDT

```txt
Field Definition Table:

   Level  I Name I Length I Format I   Options         I Flags   I Encoding
-------------------------------------------------------------------------------
  1       I  TY  I        I        I                   I         I
   2      I  TA  I    1   I    B   I DE,NU             I         I
  1       I  AA  I        I        I                   I         I
   2      I  AB  I    0   I    A   I NU                I         I
   2      I  AC  I    8   I    B   I NU                I         I
   2      I  AD  I    2   I    B   I FI                I         I
  1       I  RA  I        I        I                   I         I
   2      I  RF  I    4   I    B   I NU                I         I
   2      I  RD  I    0   I    A   I DE,NU             I         I
   2      I  RN  I    0   I    A   I DE,UQ,NU          I         I
   2      I  RB  I    1   I    B   I FI                I         I
   2      I  RO  I    0   I    A   I NU                I         I
  1       I  DL  I        I        I                   I         I
   2      I  DF  I    4   I    B   I NU                I         I
   2      I  DD  I    0   I    A   I DE,NU             I         I
  1       I  MA  I        I        I PE                I         I
   2      I  MB  I   20   I    A   I NU                I         I
   2      I  MC  I    4   I    B   I NU                I         I
   2      I  MD  I    0   I    A   I NU                I         I
   2      I  ML  I    4   I    B   I NU                I         I
   2      I  MT  I    0   I    A   I NU                I         I
   2      I  MY  I    2   I    A   I FI                I         I
   2      I  MR  I    0   I    A   I NU                I         I
  1       I  ZB  I    8   I    B   I DE,MU             I         I
          I      I        I        I DT(DATETIME)      I         I
          I      I        I        I SY=TIME           I         I
-------------------------------------------------------------------------------
```

## Usage of Adabas Maps

Because the Adabas Maps are part of the basic concept of the Java and GO API, the Adabas Maps can be used in all components. The Adabas RESTful API provides the possibility to access Adabas RESTful data using the Adabas Map name.

Inside the Adabas API the Adabas Map access is possible to be referenced using the repository and the name reference.

Because the Adabas Maps are part of the basic concept of the Java and GO API, the Adabas Maps can be used in all components. The Adabas RESTful API provides the possibility to access Adabas RESTful data using the Adabas Map name.

Inside the Adabas GO API the Adabas Map access is possible to be referenced using the repository and the name reference.

## Import and Export of maps

It may be useful to administrate the Map definition. Especially to move them from development databases to the production database.

Therefore an import/export API is introduced. The file format is JSON. Here is an example JSON configuration for a Map

```json
{"Maps":[
   {"Name":"VehicleMap",
   "Data":{
      "Target":"24(tcpip://vanGogh:0)","File":12},"LastModifified":"2019\\02\\06 20:11:26",
      "Fields":[
         {"LongName":"Vendor","ShortName":"AD","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"},
         {"LongName":"Model","ShortName":"AE","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"},
         {"LongName":"Color","ShortName":"AF","ContentType":"","Charset":"US-ASCII","File":0,"FormatType":"A","FormatLength":-1,"FieldType":"ALPHA"}
      ]
   }]
}
```

You can use the GO api to load the JSON file and write it to an Map repository like that:

```GO
maps, merr := LoadJSONMap("COPYEMPL.json")
for _, m := range maps {
  m.Repository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4}
  err = m.Store()
}
```