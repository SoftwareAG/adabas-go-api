/*
* Copyright © 2019-2020 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */

// Package adabas contains Adabas specific Adabas buffer conversion and call functions.
// The Adabas file metadata will be read and requested field content is returned.
// The package provides three type of access to the database.
//
// 1. The local access using the Adabas client native library. This uses the classic inter process
// communication method and might use the Entire Network client method accessing databases using the
// Entire Network server node infrastructure.
//
// 2. The new Adabas TCP/IP communication for a direct point-to-point access to the database. This is
// support since Adabas version v6.7.
//
// Database reference
//
// The Adabas database is referenced using a Adabas database URL. Local databases can be referenced using
// the database id, the Adabas map or a remote reference with port 0. It is possible to reference remote
// databases with the host and port directly.
//
// A local database reference: "24", "24","24(adatcp://host:0)".
//
// A remote database reference: "24(adatcp://host:123)"
//
// To use local IPC or Entire Net-Work client related Adabas access, please compile Adabas GO API with
// ADALNK library references.
// See documentation here: https://github.com/SoftwareAG/adabas-go-api
//
// Example
//
// Here a short example showing a database read accces using Adabas maps
//  connection, cerr := NewConnection("acj;map;config=[24,4]")
//  if cerr != nil {
//  	return cerr
//  }
//  defer connection.Close()
//  request, rerr := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
//  if rerr != nil {
//  	fmt.Println("Error create request", rerr)
//  	return rerr
//  }
//  err := request.QueryFields("NAME,FIRST-NAME,PERSONNEL-ID")
//  if !assert.NoError(b, err) {
//  	return err
//  }
//  request.Limit = 0
//  result, rErr := request.ReadLogicalBy("NAME")
//  result.DumpValues()
//
// Read logic
//
// You may read using search values and descriptor sorted searches.
// The received records can be analyzed using traversation logic.
// See documentation here:  https://github.com/SoftwareAG/adabas-go-api/blob/master/doc/QUERY.md
//
// Adabas maps
//
// For long name and database name usage, a new Adabas map concept is introduced. The Adabas maps
// are stored inside the database.
// See documentation here: https://github.com/SoftwareAG/adabas-go-api/blob/master/doc/AdabasMap.md
//
// Stream
//
// It is possible to work with the records just-in-time they received
// in a stream. A callback function will be called to process the current
// received record.
//
// Example using stream:
//   func dumpStream(record *Record, x interface{}) error {
//  	i := x.(*uint32)
//  	a, _ := record.SearchValue("AE")
//  	fmt.Printf("Read %d -> %s = %d\n", record.Isn, a, record.Quantity)
//  	(*i)++
//   	return nil
//   }
//   result, err := request.ReadLogicalWithStream("AE='SMITH'", dumpStream, &i)
//
// Struct usage
//
// Example for a structure:
//   type FullName struct {
//      FirstName string
//      LastName  string
//   }
//   type EmployeeMap struct {
//      ID         string `adabas:"Id:key"`
//      Name       *FullName
//      Department []byte
//   }
//
// For a Adabas Map called EmployeeMap you can use the structure to read and write
// Adabas data into the database.
//
// Example to read Adabas data using the structure:
//   request, rerr := connection.CreateMapReadRequest((*EmployeeMap)(nil))
//   err := request.QueryFields("FullName")
//   result, err = request.ReadLogicalBy("LastName")
//
// The result list of structure entries will be in the result. You can reference
// the list using the result.Data list.
// If using the stream callback method, then the stream will get an EmployeesMap
// instance instead of a Record instance.
//
// Partial large objects
//
// You can read a large objects using the `ReadStream` method to subdivide a
// big large object into slices reading parts of the large objects instead of
// read one big record of the large object.
// The blocksize read in one stream call is defined in the `ReadRequest.Blocksize`.
//
package adabas
