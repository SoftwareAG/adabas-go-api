package adabas

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type FullName struct {
	FirstName string
	LastName  string
}

type Income struct {
	Salary   uint64
	Bonus    []uint64
	Currency string
}

type EmployeesSalary struct {
	ID         string `adabas:"Id"`
	FullName   *FullName
	Birth      uint64
	Department string
	Income     []*Income
	Language   []string
}

func TestStoreRequestInterfaceInstance(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(Employees{}, ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "Employees", storeRequest.dynamic.DataType.Name())
}

func TestStoreRequestInterfacePointer(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest((*Employees)(nil), ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "Employees", storeRequest.dynamic.DataType.Name())
}

func refreshFile(modDbid string, fnr Fnr) error {
	connection, err := NewConnection("ada;target=" + modDbid)
	if err != nil {
		return err
	}
	defer connection.Close()
	fmt.Println(connection)
	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(fnr)
	if rErr != nil {
		return rErr
	}
	readRequest.QueryFields("")
	deleteRequest, dErr := connection.CreateDeleteRequest(fnr)
	if dErr != nil {
		return dErr
	}
	readRequest.Limit = 0
	err = readRequest.ReadPhysicalSequenceWithParser(deleteRecords, deleteRequest)
	if err != nil {
		return err
	}
	err = deleteRequest.EndTransaction()
	if err != nil {
		return err
	}
	return nil
}

func TestStoreInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	rerr := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, rerr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(Employees{}, ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "Employees", storeRequest.dynamic.DataType.Name())

	employees := make([]*Employees, 0)
	employees = append(employees, &Employees{ID: "ID", Birth: 123, Name: "Name", FirstName: "First name"})
	employees = append(employees, &Employees{ID: "ID2", Birth: 234, Name: "Name2", FirstName: "First name2"})
	employees = append(employees, &Employees{ID: "ABC", Birth: 978, Name: "XXX", FirstName: "HHHH name"})
	err = storeRequest.StoreData(employees)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.StoreData(&Employees{ID: "ID3", Birth: 456, Name: "Name3", FirstName: "First name3"})
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.StoreData(Employees{ID: "ID4", Birth: 789, Name: "Name4", FirstName: "First name4"})
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("End transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

}

func TestStorePeriodInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	rerr := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, rerr) {
		return
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(EmployeesSalary{}, ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "EmployeesSalary", storeRequest.dynamic.DataType.Name())

	employees := make([]*EmployeesSalary, 0)
	income := make([]*Income, 0)
	income = append(income, &Income{Currency: "EUR", Salary: 40000, Bonus: []uint64{123, 123}})
	income = append(income, &Income{Currency: "EUR", Salary: 60000, Bonus: []uint64{1000, 1500}})
	employees = append(employees, &EmployeesSalary{ID: "pId123", Birth: 123344,
		FullName: &FullName{LastName: "Overmeyer", FirstName: "Ottofried"}, Department: "FBI",
		Income: income, Language: []string{"ENG", "FRA"}})
	income = make([]*Income, 0)
	income = append(income, &Income{Currency: "LIR", Salary: 400000, Bonus: []uint64{40000, 5000}})
	income = append(income, &Income{Currency: "PFD", Salary: 6000000, Bonus: []uint64{100000, 10000000}})
	employees = append(employees, &EmployeesSalary{ID: "pId007", Birth: 5555555,
		FullName: &FullName{LastName: "Bond", FirstName: "James"}, Department: "MI5",
		Income: income, Language: []string{"ENG", "FRA", "GER", "MAN"}})
	err = storeRequest.StoreData(employees)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("End transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return
	}

}

func TestReadLogicalInterface(t *testing.T) {
	err := initLogWithFile("request_interface.log")
	if err != nil {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	// err = request.QueryFields("*")
	// if !assert.NoError(t, err) {
	// 	return
	// }
	assert.Equal(t, "Employees", request.dynamic.DataType.Name())

	result, err := request.ReadLogicalWith("ID=['ID':'ID9']")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
		result.DumpData()
		assert.Len(t, result.Data, 4)
		e := result.Data[0].(*Employees)
		assert.Equal(t, "ID", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name", strings.Trim(e.Name, " "))
		e = result.Data[1].(*Employees)
		assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name2", strings.Trim(e.Name, " "))
		e = result.Data[2].(*Employees)
		assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name3", strings.Trim(e.Name, " "))
		e = result.Data[3].(*Employees)
		assert.Equal(t, "ID4", strings.Trim(e.ID, " "))
		assert.Equal(t, int64(789), e.Birth)
		assert.Equal(t, "Name4", strings.Trim(e.Name, " "))
	}
}

func TestReadPhysicalInterface(t *testing.T) {
	err := initLogWithFile("request_interface.log")
	if err != nil {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}

	result, err := request.ReadPhysicalSequence()
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
		result.DumpData()
		fmt.Println("Length", len(result.Data))
		assert.True(t, len(result.Data) > 4)
		if assert.IsType(t, (*Employees)(nil), result.Data[0]) {
			nrNotFound := 3
			for _, x := range result.Data {
				e := x.(*Employees)
				switch {
				case strings.HasPrefix(e.ID, "ID "):
					assert.Equal(t, "ID", strings.Trim(e.ID, " "))
					assert.Equal(t, int64(123), e.Birth)
					assert.Equal(t, "Name", strings.Trim(e.Name, " "))
					nrNotFound--
				case strings.HasPrefix(e.ID, "ID2 "):
					assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
					assert.Equal(t, "Name2", strings.Trim(e.Name, " "))
					nrNotFound--
				case strings.HasPrefix(e.ID, "ID3 "):
					assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
					assert.Equal(t, "Name3", strings.Trim(e.Name, " "))
					nrNotFound--
				default:
				}

			}
			assert.Equal(t, 0, nrNotFound)
		}
	}
}

func receiveInterface(data interface{}, x interface{}) error {
	i := x.(*int)
	*i++
	e := data.(*Employees)
	if strings.HasSuffix(e.ID, "ID") {
		fmt.Println("Error data incorrect ....")
		return fmt.Errorf("Error data incorrect")
	}
	fmt.Println(data)
	return nil
}

func TestReadLogicalInterfaceStream(t *testing.T) {
	err := initLogWithFile("request_interface.log")
	if err != nil {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return
	}

	i := 0
	result, err := request.ReadLogicalWithInterface("ID=['ID':'ID9']", receiveInterface, &i)
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 4, i)
	assert.Nil(t, result.Values)
	assert.Nil(t, result.Data)
}

func TestReadLogicalPeriodInterface(t *testing.T) {
	err := initLogWithFile("request_interface.log")
	if err != nil {
		return
	}

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(EmployeesSalary{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	// err = request.QueryFields("*")
	// if !assert.NoError(t, err) {
	// 	return
	// }
	assert.Equal(t, "EmployeesSalary", request.dynamic.DataType.Name())

	result, err := request.ReadLogicalWith("Id=['pId':'pId9']")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if !assert.NotNil(t, result) {
		return
	}
	result.DumpValues()
	result.DumpData()
	if !assert.Len(t, result.Data, 2) {
		return
	}
	e := result.Data[0].(*EmployeesSalary)
	assert.Equal(t, "pId007", strings.Trim(e.ID, " "))
	if !assert.NotNil(t, e.FullName) {
		return
	}
	assert.Equal(t, "Bond", strings.Trim(e.FullName.LastName, " "))
	e = result.Data[1].(*EmployeesSalary)
	assert.Equal(t, "pId123", strings.Trim(e.ID, " "))
	assert.Equal(t, "Overmeyer", strings.Trim(e.FullName.LastName, " "))
	assert.Equal(t, uint64(123344), e.Birth)
	assert.Equal(t, "FBI   ", e.Department)
	if assert.Len(t, e.Language, 2) {
		assert.Equal(t, "ENG", e.Language[0])
	}
	if assert.Len(t, e.Income, 2) {
		assert.Equal(t, uint64(40000), e.Income[0].Salary)
		if assert.Len(t, e.Income[0].Bonus, 2) {
			assert.Equal(t, uint64(123), e.Income[0].Bonus[0])
		}

		assert.Equal(t, "EUR", e.Income[0].Currency)
	}

}

func TestRequestLogicalByEmployeesSalary(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	adabas, _ := NewAdabas(23)
	request, err := NewReadRequest("EmployeesSalary", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return
	}
	defer request.Close()
	_, openErr := request.Open()
	if assert.NoError(t, openErr) {
		err = request.QueryFields("Id, FullName, FirstName, LastName, MiddleName, Birth, Telephone, AreaCode, Phone,Department, Income, Currency, Salary, Bonus,Language")
		if err != nil {
			return
		}
		result, err := request.ReadLogicalWith("Id=['pId':'pId9']")
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			result.DumpValues()
		}
	}
}
