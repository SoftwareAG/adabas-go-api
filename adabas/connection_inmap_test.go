/*
* Copyright © 2020 Software AG, Darmstadt, Germany and/or its licensors
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

package adabas

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type IncomeInMap struct {
	Salary   uint64   `adabas:"::AS"`
	Bonus    []uint64 `adabas:"::AT"`
	Currency string   `adabas:"::AR"`
	Summary  uint64   `adabas:":ignore"`
}

type EmployeesInMap struct {
	Index      uint64         `adabas:":isn"`
	ID         string         `adabas:":key:AA"`
	FullName   *FullNameInMap `adabas:"::AB"`
	Birth      uint64         `adabas:"::AH"`
	Department string         `adabas:"::AO"`
	Income     []*IncomeInMap `adabas:"::AQ"`
	Language   []string       `adabas:"::AZ"`
}

type FullNameInMap struct {
	FirstName  string `adabas:"::AC"`
	MiddleName string `adabas:"::AD"`
	Name       string `adabas:"::AE"`
}

type NewEmployeesInMap struct {
	Index  uint64                `adabas:":isn"`
	ID     string                `adabas:":key:AA"`
	Income []*NewEmployeesIncome `adabas:"::L0"`
}

type NewEmployeesIncome struct {
	CurCode string `adabas:"::LA"`
	Salary  int    `adabas:"::LB"`
	Bonus   []int  `adabas:"::LC"`
}

type VehicleInMap struct {
	Index uint64 `adabas:":isn"`
	Reg   string `adabas:":key:AA"`
	ID    string `adabas:"::AC"`
	Model string `adabas:"::AE"`
	Color string `adabas:"::AF"`
	Year  uint64 `adabas:"::AG"`
}

func TestInlineMap(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23,11")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.ReadISN(1024)
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	if assert.Len(t, response.Data, 1) {
		assert.Equal(t, "30021228", response.Data[0].(*EmployeesInMap).ID)
		assert.Equal(t, "JAMES               ", response.Data[0].(*EmployeesInMap).FullName.FirstName)
		assert.Equal(t, "SMEDLEY             ", response.Data[0].(*EmployeesInMap).FullName.Name)
		assert.Equal(t, "COMP02", response.Data[0].(*EmployeesInMap).Department)
	}
	_ = response.DumpValues()
	assert.Len(t, response.Values, 0)
}

func TestInlineMapSearchAndOrder(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23,11")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.SearchAndOrder("AA=50005600", "AE")
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	if assert.Len(t, response.Data, 1) {
		assert.Equal(t, "50005600", response.Data[0].(*EmployeesInMap).ID)
		assert.Equal(t, "HUMBERTO            ", response.Data[0].(*EmployeesInMap).FullName.FirstName)
		assert.Equal(t, "MORENO              ", response.Data[0].(*EmployeesInMap).FullName.Name)
		assert.Equal(t, "VENT07", response.Data[0].(*EmployeesInMap).Department)
	}
	_ = response.DumpValues()
	assert.Len(t, response.Values, 0)
}

func TestInlineMapHistogram(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23,11")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.HistogramWith("AO=VENT07")
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	assert.Len(t, response.Data, 0)
	_ = response.DumpValues()
	if assert.Len(t, response.Values, 1) {
		assert.Equal(t, uint64(5), response.Values[0].Quantity)
	}
}

func TestInlineMapHistogramDesc(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23,11")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 4
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.HistogramBy("AO")
	if !assert.NoError(t, rerr) {
		return
	}
	_ = response.DumpData()
	assert.Len(t, response.Data, 0)
	_ = response.DumpValues()
	if assert.Len(t, response.Values, 4) {
		assert.Equal(t, uint64(5), response.Values[0].Quantity)
		assert.Equal(t, "ADMA01", response.Values[0].HashFields["AO"].String())
		assert.Equal(t, uint64(35), response.Values[3].Quantity)
		assert.Equal(t, "COMP02", response.Values[3].HashFields["AO"].String())
	}
}

func TestInlineStoreMap(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	_ = clearAdabasFile(t, adabasModDBIDs, 16)

	fmt.Println("Starting inmap store ....")
	connection, cerr := NewConnection("acj;inmap=23,16")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapStoreRequest(&EmployeesInMap{})
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Storing fields ....")
	err = request.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	e := &EmployeesInMap{FullName: &FullNameInMap{FirstName: "Anton", Name: "Skeleton", MiddleName: "Otto"}, Birth: 1234}
	fmt.Println("Storing record ....")
	rerr := request.StoreData(e)
	if !assert.NoError(t, rerr) {
		return
	}
	err = request.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "inmapstore", "23", 16)
}

type EmployeeInMapPe struct {
	Name        string         `adabas:"::AE"`
	FirstName   string         `adabas:"::AD"`
	Isn         uint64         `adabas:":isn"`
	ID          string         `adabas:"::AA"`
	Income      []*InmapIncome `adabas:"::AQ"`
	AddressLine []string       `adabas:"::AI"`
	LeaveBooked []*InmapLeave  `adabas:"::AW"`
}

// Income income
type InmapIncome struct {
	Currency string `adabas:"::AR"`
	Salary   uint32 `adabas:"::AS"`
}

type InmapLeave struct {
	LeaveStart uint64 `adabas:"::AX"`
	LeaveEnd   uint64 `adabas:"::AY"`
}

type names string
type birth int
type EmployeeInMapType struct {
	ID        string `adabas:"::AA"`
	Name      names  `adabas:"::AE"`
	FirstName names  `adabas:"::AD"`
	Birth     birth  `adabas:"::AH"`
	Isn       uint64 `adabas:":isn"`
}

func TestInlineStorePE(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	clearAdabasFile(t, adabasModDBIDs, 16)

	fmt.Println("Starting inmap store ....")
	connection, cerr := NewConnection("acj;inmap=23,16")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapStoreRequest(&EmployeeInMapPe{})
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Storing fields ....")
	err = request.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	j := &EmployeeInMapPe{Name: "XXX", ID: "fdlldnfg", LeaveBooked: []*InmapLeave{{LeaveStart: 3434, LeaveEnd: 232323}}}
	fmt.Println("Storing record ....")
	rerr := request.StoreData(j)
	if !assert.NoError(t, rerr) {
		return
	}
	err = request.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "inmapstorepe", "23", 16)
}

func TestInlineStorePEMU(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	clearAdabasFile(t, adabasModDBIDs, 16)

	fmt.Println("Starting inmap store ....")
	connection, cerr := NewConnection("acj;inmap=23,16")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapStoreRequest(&EmployeeInMapPe{})
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Storing fields ....")
	err = request.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	j := &EmployeeInMapPe{Name: "XXX", ID: "fdlldnfg", Income: []*InmapIncome{{Currency: "ABB", Salary: 121324}}}
	fmt.Println("Storing record ....")
	rerr := request.StoreData(j)
	if !assert.NoError(t, rerr) {
		return
	}
	err = request.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}
	checkContent(t, "inmapstorepemu", "23", 16)
}

type Executions struct {
	Isn        uint64 `adabas:":isn"`
	ID         int64  `json:"Id" adabas:"::"`
	User       string `adabas:"::US"`
	HashID     string `adabas:"::JI"`
	Scheduled  int64  `adabas:"::SB"`
	Ended      int64  `adabas:"::SE"`
	ExitCode   int    `adabas:"::EX"`
	RecordType byte   `adabas:"::TY"`
	Flags      byte   `adabas:"::JL"`
	StartedBy  string `adabas:"::NA"`
	ChangeTime string `adabas:"::ZB"`
	LogFile    string `json:"Log" adabas:"::"`
	LogContent string `json:"-" adabas:"::LO"`
	JobDesc    string `adabas:"::QJ"`
}

const logText = `%ADAREP-I-STARTED,      02-MAR-2021 21:57:34, Version 7.1.0.0 (MacOSX 64Bit)
%ADAREP-F-DBSLMM, Structure level mismatch on database
%ADAREP-E-ACLOGPERM, The permissions of the LOG_FILE directory are insufficient.
%ADAREP-I-IOCNT, 1 IOs on dataset ASSO
%ADAREP-I-ABORTED,      02-MAR-2021 21:57:34, elapsed time: 00:00:00
`

func TestInmapExecutions(t *testing.T) {
	initTestLogWithFile(t, "inmap.log")
	connection, cerr := NewConnection("acj;inmap=23,5")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	read, err := connection.CreateMapReadRequest(&Executions{})
	if !assert.NoError(t, err) {
		return
	}
	result, rErr := read.SearchAndOrder("QJ=Eef60560cf41eb28856eb61544ef04f2b", "QJ")
	if !assert.NoError(t, rErr) {
		return
	}
	assert.Len(t, result.Data, 1)
	e := result.Data[0].(*Executions)
	assert.Equal(t, logText, e.LogContent)

	writeJobEx, err := connection.CreateMapStoreRequest(&Executions{})
	if !assert.NoError(t, err) {
		return
	}
	jobEx := &Executions{ID: 1223224, User: "TestUser",
		HashID: "0fdeaa0022", LogFile: "testfile.notsee",
		LogContent: "fndlsnfsldfnsldfnsldfnsödfns.dfnsdfnsldfnslfnsölfnsl"}
	err = writeJobEx.StoreData(jobEx)
	if !assert.NoError(t, err) {
		return
	}
}

func TestInmapEmployeesRetyped(t *testing.T) {
	initTestLogWithFile(t, "inmap.log")

	clearAdabasFile(t, adabasModDBIDs, 16)

	connection, cerr := NewConnection("acj;inmap=23,16")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	x := &EmployeeInMapType{Name: "abc", FirstName: "Otto",
		ID: "AA123", Birth: 12345}
	store, err := connection.CreateMapStoreRequest(x)
	if !assert.NoError(t, err) {
		return
	}
	err = store.StoreData(x)
	if !assert.NoError(t, err) {
		return
	}
	err = store.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

	read, err := connection.CreateMapReadRequest(&EmployeeInMapType{})
	if !assert.NoError(t, err) {
		return
	}
	result, rErr := read.SearchAndOrder("AA="+x.ID, "AA")
	if !assert.NoError(t, rErr) {
		return
	}

	if assert.Len(t, result.Data, 1) {
		e := result.Data[0].(*EmployeeInMapType)
		assert.Equal(t, "AA123   ", e.ID)
		assert.Equal(t, names("abc                 "), e.Name)
		assert.Equal(t, names("Otto                "), e.FirstName)
		assert.Equal(t, birth(12345), e.Birth)
	}
}

func TestInlineMapPeriodSearchAndOrder(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=24,9")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest((*NewEmployeesInMap)(nil))
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.SearchAndOrder("AA=11100108", "AA")
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.Len(t, response.Data, 1) {
		entry := response.Data[0].(*NewEmployeesInMap)
		assert.Equal(t, "11100108", entry.ID)
		assert.Equal(t, uint64(208), entry.Index)
		assert.Len(t, entry.Income, 4)
		x := []string{"EUR", "EUR", "EUR", "EUR"}
		y := []int{22564, 21538, 20000, 18974}
		z := [][]int{[]int{1538}, []int{}, []int{}, []int{}}
		for i, e := range entry.Income {
			assert.Equal(t, x[i], e.CurCode)
			assert.Equal(t, y[i], e.Salary)
			assert.NotNil(t, e.Bonus)
			assert.Len(t, e.Bonus, len(z[i]), fmt.Sprintf("Index %d wrong %v", i, e.Bonus))
			assert.Equal(t, z[i], e.Bonus)
		}
	}
	assert.Len(t, response.Values, 0)
	response, rerr = request.SearchAndOrder("AA=11300321", "AA")
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.Len(t, response.Data, 1) {
		entry := response.Data[0].(*NewEmployeesInMap)
		assert.Equal(t, "11300321", entry.ID)
		assert.Len(t, entry.Income, 5)
	}
}

type Parameter struct {
	Parameter string `adabas:"::PN"`
}

type Job struct {
	Isn        uint64      `adabas:":isn"`
	Name       string      `adabas:"Id:key:NA"`
	HashID     string      `adabas:"::JI"`
	Flags      byte        `adabas:"::JL"`
	User       string      `adabas:"::US"`
	Parameters []Parameter `adabas:"::PA"`
}

func TestInlineMapJobSearchAndOrder(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOARCH == "arm" {
		t.Skip("Not supported on this architecture")
		return
	}
	initTestLogWithFile(t, "inmap.log")

	connection, cerr := NewConnection("acj;inmap=23")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()
	adatypes.Central.Log.Debugf("Created connection : %#v", connection)
	request, err := connection.CreateMapReadRequest((*Job)(nil), 5)
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}
	response, rerr := request.SearchAndOrder("TY=J", "NA")
	if !assert.NoError(t, rerr) {
		return
	}
	if assert.Len(t, response.Data, 5) {
		entry := response.Data[0].(*Job)
		assert.Equal(t, "ADAREP", entry.Name)
		assert.Equal(t, uint8(0), entry.Flags)
		assert.Len(t, entry.Parameters, 2)
		assert.Equal(t, Parameter(Parameter{Parameter: "db=24"}), entry.Parameters[0])
		assert.Equal(t, Parameter(Parameter{Parameter: "ET_SYNC_WAIT=10"}), entry.Parameters[1])
	}
}
