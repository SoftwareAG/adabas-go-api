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

type EmployeesKey struct {
	ID         string `adabas:"Id:key"`
	FullName   *FullName
	Birth      string
	Department string
	Language   []string
}

type EmployeesIndex struct {
	Index    uint64 `adabas:":isn"`
	ID       string `adabas:"Id:key"`
	LastName string
}

func TestStoreRequestInterfaceInstance(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)

	_, err := repository.SearchMap(ada, "EmployeesSalary")
	if err != nil {
		maps, merr := LoadJSONMap("EmployeesSalary.json")
		if !assert.NoError(t, merr) && !assert.Len(t, maps, 1) {
			return
		}
		maps[0].Repository = &repository.DatabaseURL
		err = maps[0].Store()
		if !assert.NoError(t, err) {
			return
		}
	}

	_, err = repository.SearchMap(ada, "EmployeesKey")
	if err != nil {
		maps, merr := LoadJSONMap("EmployeesKey.json")
		if !assert.NoError(t, merr) && !assert.Len(t, maps, 1) {
			return
		}
		maps[0].Repository = &repository.DatabaseURL
		err = maps[0].Store()
		if !assert.NoError(t, err) {
			return
		}
	}

	_, err = repository.SearchMap(ada, "EmployeesIndex")
	if err != nil {
		maps, merr := LoadJSONMap("EmployeesIndex.json")
		if !assert.NoError(t, merr) && !assert.Len(t, maps, 1) {
			return
		}
		maps[0].Repository = &repository.DatabaseURL
		err = maps[0].Store()
		if !assert.NoError(t, err) {
			return
		}
	}

	storeRequest, err := NewStoreRequest(Employees{}, ada, repository)
	if !assert.NoError(t, err) {
		return
	}
	err = storeRequest.StoreFields("*")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotEqual(t, (*adatypes.DynamicInterface)(nil), storeRequest.commonRequest.dynamic)
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "Employees", storeRequest.dynamic.DataType.Name())
	readRequest, rErr := NewReadRequest(storeRequest)
	if !assert.NoError(t, rErr) {
		return
	}
	assert.NotEqual(t, &storeRequest.commonRequest, &readRequest.commonRequest)
	assert.Equal(t, (*adatypes.DynamicInterface)(nil), readRequest.commonRequest.dynamic)
	assert.NotEqual(t, (*adatypes.DynamicInterface)(nil), storeRequest.commonRequest.dynamic)
	assert.NotEqual(t, storeRequest.commonRequest.definition, readRequest.commonRequest.definition)
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
	if storeInterface(t) != nil {
		return
	}
	if readLogicalInterface(t) != nil {
		return
	}
	if readPhysicalInterface(t) != nil {
		return
	}
	if readLogicalInterfaceStream(t) != nil {
		return
	}
	if readLogicalIndexInterface(t) != nil {
		return
	}
}

func storeInterface(t *testing.T) error {
	rerr := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, rerr) {
		return rerr
	}
	fmt.Println("Store interface")
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(Employees{}, ada, repository)
	if !assert.NoError(t, err) {
		return err
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "Employees", storeRequest.dynamic.DataType.Name())

	employees := make([]*Employees, 0)
	employees = append(employees, &Employees{ID: "ID", Birth: 711999, Name: "Name", FirstName: "First name"})
	employees = append(employees, &Employees{ID: "ID2", Birth: 234, Name: "Name2", FirstName: "First name2"})
	employees = append(employees, &Employees{ID: "ABC", Birth: 978, Name: "XXX", FirstName: "HHHH name"})
	err = storeRequest.StoreData(employees)
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRequest.StoreData(&Employees{ID: "ID3", Birth: 456, Name: "Name3", FirstName: "First name3"})
	if !assert.NoError(t, err) {
		return err
	}
	err = storeRequest.StoreData(Employees{ID: "ID4", Birth: 711714, Name: "Name4", FirstName: "First name4"})
	if !assert.NoError(t, err) {
		return err
	}
	fmt.Println("End transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return err
	}
	return nil
}

func updateKeyInterface(t *testing.T) error {
	fmt.Println("Update interface records")
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(EmployeesKey{}, ada, repository)
	if !assert.NoError(t, err) {
		return err
	}
	assert.NotNil(t, storeRequest)
	assert.Equal(t, "EmployeesKey", storeRequest.dynamic.DataType.Name())

	employees := make([]*EmployeesKey, 0)
	employees = append(employees, &EmployeesKey{ID: "ID", FullName: &FullName{LastName: "NewName", FirstName: "First name"}})
	employees = append(employees, &EmployeesKey{ID: "ID2", FullName: &FullName{LastName: "NewName2", FirstName: "First name2"}})
	employees = append(employees, &EmployeesKey{ID: "ID4", Birth: "2012/10/30", FullName: &FullName{LastName: "ZZZZZZZ", FirstName: "UUUUUU name"}})
	err = storeRequest.UpdateData(employees)
	if !assert.NoError(t, err) {
		return err
	}
	fmt.Println("End transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return err
	}
	return nil
}

func readLogicalInterface(t *testing.T) error {
	fmt.Println("Read logical interface")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
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
		return err
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
		result.DumpData()
		assert.Len(t, result.Data, 4)
		e := result.Data[0].(*Employees)
		assert.Equal(t, "ID", strings.Trim(e.ID, " "))
		if !assert.Equal(t, "Name", strings.Trim(e.Name, " ")) {
			return fmt.Errorf("Name mismatch")
		}
		e = result.Data[1].(*Employees)
		assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name2", strings.Trim(e.Name, " "))
		e = result.Data[2].(*Employees)
		assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name3", strings.Trim(e.Name, " "))
		e = result.Data[3].(*Employees)
		assert.Equal(t, "ID4", strings.Trim(e.ID, " "))
		assert.Equal(t, int64(711714), e.Birth)
		assert.Equal(t, "Name4", strings.Trim(e.Name, " "))
	}
	return nil
}

func readLogicalIndexInterface(t *testing.T) error {
	fmt.Println("Read logical index to interface")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(EmployeesIndex{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	// err = request.QueryFields("*")
	// if !assert.NoError(t, err) {
	// 	return
	// }
	assert.Equal(t, "EmployeesIndex", request.dynamic.DataType.Name())

	result, err := request.ReadLogicalWith("Id=['ID':'ID9']")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return err
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
		result.DumpData()
		assert.Len(t, result.Data, 4)
		e := result.Data[0].(*EmployeesIndex)
		assert.True(t, e.Index > 0)
		assert.Equal(t, "ID", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name", strings.Trim(e.LastName, " "))
		e = result.Data[1].(*EmployeesIndex)
		assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
		e = result.Data[2].(*EmployeesIndex)
		assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
		e = result.Data[3].(*EmployeesIndex)
		assert.Equal(t, "ID4", strings.Trim(e.ID, " "))
		assert.NotEqual(t, uint64(0), e.Index)
		assert.Equal(t, "Name4", strings.Trim(e.LastName, " "))
	}

	storeRequest, err := NewStoreRequest(EmployeesIndex{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer storeRequest.Close()
	e := result.Data[0].(*EmployeesIndex)
	fmt.Printf("Update record on ISN=%d with Id= %s and last name=%s\n",
		e.Index, e.ID, e.LastName)
	adatypes.Central.Log.Debugf("TEST: Update record on ISN=%d with Id= %s and last name=%s\n",
		e.Index, e.ID, e.LastName)
	e.LastName = "updateindexname"
	err = storeRequest.UpdateData(e)
	if assert.NoError(t, err) {
		adatypes.Central.Log.Debugf("TEST: Update done on ISN=%d\n",
			e.Index)
		err = storeRequest.EndTransaction()
		assert.NoError(t, err)
	}
	request, err = NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	result, err = request.ReadLogicalWith("ID='ID'")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return err
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		e := result.Data[0].(*Employees)
		assert.Equal(t, "ID", strings.Trim(e.ID, " "))
		assert.Equal(t, "updateindexname", strings.Trim(e.Name, " "))
		assert.Equal(t, "First name", strings.Trim(e.FirstName, " "))
	}
	return nil
}

func readPhysicalInterface(t *testing.T) error {
	fmt.Println("Read physical interface")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return err
	}

	result, err := request.ReadPhysicalSequence()
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return err
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
					assert.Equal(t, int64(711999), e.Birth)
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
	return nil
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

func readLogicalInterfaceStream(t *testing.T) error {
	fmt.Println("Read logical interface stream")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(Employees{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	err = request.QueryFields("*")
	if !assert.NoError(t, err) {
		return err
	}

	i := 0
	result, err := request.ReadLogicalWithInterface("ID=['ID':'ID9']", receiveInterface, &i)
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return err
	}
	assert.Equal(t, 4, i)
	assert.Nil(t, result.Values)
	assert.Nil(t, result.Data)
	return nil
}

func TestStorePeriodInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	if storePeriodInterface(t) != nil {
		return
	}
	if readLogicalPeriodInterface(t) != nil {
		return
	}
	if readLogicalPeriodInterfaceByEmployeesSalary(t) != nil {
		return
	}
}

func storePeriodInterface(t *testing.T) error {
	fmt.Println("Store interface with period group")

	rerr := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, rerr) {
		return rerr
	}
	ada, _ := NewAdabas(adabasModDBID)
	defer ada.Close()
	repository := NewMapRepository(ada, 4)
	storeRequest, err := NewStoreRequest(EmployeesSalary{}, ada, repository)
	if !assert.NoError(t, err) {
		return err
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
		return err
	}
	fmt.Println("End transaction")
	err = storeRequest.EndTransaction()
	if !assert.NoError(t, err) {
		return err
	}
	return nil
}

func readLogicalPeriodInterface(t *testing.T) error {
	fmt.Println("Read interface with period group")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(EmployeesSalary{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
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
		return err
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if !assert.NotNil(t, result) {
		return fmt.Errorf("Error got")
	}
	result.DumpValues()
	result.DumpData()
	if !assert.Len(t, result.Data, 2) {
		return fmt.Errorf("Error got")
	}
	e := result.Data[0].(*EmployeesSalary)
	assert.Equal(t, "pId007", strings.Trim(e.ID, " "))
	if !assert.NotNil(t, e.FullName) {
		return fmt.Errorf("Error got")
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
	return nil
}

func readLogicalPeriodInterfaceByEmployeesSalary(t *testing.T) error {
	fmt.Println("Read interface with period group with EmployeesSalary")

	adabas, _ := NewAdabas(23)
	request, err := NewReadRequest("EmployeesSalary", adabas,
		NewMapRepository(adabas, 4))
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	_, openErr := request.Open()
	if assert.NoError(t, openErr) {
		err = request.QueryFields("Id, FullName, FirstName, LastName, MiddleName, Birth, Telephone, AreaCode, Phone,Department, Income, Currency, Salary, Bonus,Language")
		if err != nil {
			return err
		}
		result, err := request.ReadLogicalWith("Id=['pId':'pId9']")
		assert.NoError(t, err)
		if assert.NotNil(t, result) {
			result.DumpValues()
		}
	}
	return nil
}

func verifyUpdateLogicalInterface(t *testing.T) error {
	fmt.Println("Verify update with logical interface")

	adabas, _ := NewAdabas(23)
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(EmployeesKey{}, adabas, mapRepository)
	if !assert.NoError(t, err) {
		return err
	}
	defer request.Close()
	// err = request.QueryFields("*")
	// if !assert.NoError(t, err) {
	// 	return
	// }
	assert.Equal(t, "EmployeesKey", request.dynamic.DataType.Name())

	result, err := request.ReadLogicalWith("Id=['ID':'ID9']")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return err
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
		result.DumpData()
		assert.Len(t, result.Data, 4)
		e := result.Data[0].(*EmployeesKey)
		assert.Equal(t, "ID", strings.Trim(e.ID, " "))
		if !assert.Equal(t, "NewName", strings.Trim(e.FullName.LastName, " ")) {
			return fmt.Errorf("Name mismatch")
		}
		e = result.Data[1].(*EmployeesKey)
		assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
		assert.Equal(t, "NewName2", strings.Trim(e.FullName.LastName, " "))
		e = result.Data[2].(*EmployeesKey)
		assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
		assert.Equal(t, "Name3", strings.Trim(e.FullName.LastName, " "))
		e = result.Data[3].(*EmployeesKey)
		assert.Equal(t, "ID4", strings.Trim(e.ID, " "))
		assert.Equal(t, string("2012/10/30"), e.Birth)
		assert.Equal(t, "ZZZZZZZ", strings.Trim(e.FullName.LastName, " "))
	}
	return nil
}

func TestStoreKeyInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	if storeInterface(t) != nil {
		return
	}
	{
		adabas, _ := NewAdabas(adabasModDBID)
		request, _ := NewReadRequest(adabas, 16)
		defer request.Close()
		request.QueryFields("")
		result, err := request.ReadLogicalWith("AA=='ID      '")
		fmt.Println("Dump result received ...", err)
		if !assert.NoError(t, err) {
			return
		}
		if result != nil {
			result.DumpValues()
		}
	}

	if updateKeyInterface(t) != nil {
		return
	}
	if verifyUpdateLogicalInterface(t) != nil {
		return
	}
}

func TestCheckRead(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, _ := NewAdabas(23)
	defer adabas.Close()
	mapRepository := NewMapRepository(adabas, 4)
	request, err := NewReadRequest(EmployeesKey{}, adabas, mapRepository)
	//	request, err := NewReadRequest("EmployeesKey", adabas, mapRepository)
	if !assert.NoError(t, err) {
		return
	}
	err = request.QueryFields("")
	if !assert.NoError(t, err) {
		return
	}
	//	request.QueryFields("Id")
	result, rErr := request.ReadLogicalBy("Id")
	if !assert.NoError(t, rErr) {
		return
	}
	fmt.Println("Length evaluating ISN", len(result.Values), len(result.Data))
	assert.Len(t, result.Values, 0)
	assert.Greater(t, len(result.Data), 0)
}

func TestConnectionStoreUsingInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	rerr := refreshFile(adabasModDBIDs, 16)
	if !assert.NoError(t, rerr) {
		return
	}

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapStoreRequest(Employees{})
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err := request.StoreFields("ID,Name")
	if !assert.NoError(t, err) {
		return
	}
	request.definition.DumpTypes(false, true, "Active")
	e := Employees{ID: "CONTEST", Name: "ConnectionTest"}
	err = request.StoreData(e)
	if !assert.NoError(t, err) {
		return
	}
	e = Employees{ID: "CONTEST2", Name: "SecondConnectionTest"}
	err = request.StoreData(e)
	if !assert.NoError(t, err) {
		return
	}
	err = connection.EndTransaction()
	assert.NoError(t, err)
}

func TestConnectionUsingInterface(t *testing.T) {
	initTestLogWithFile(t, "request_interface.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())

	connection, cerr := NewConnection("acj;map;config=[" + adabasModDBIDs + ",4]")
	if !assert.NoError(t, cerr) {
		return
	}
	defer connection.Close()

	request, rerr := connection.CreateMapReadRequest(Employees{})
	if !assert.NoError(t, rerr) {
		fmt.Println("Error create request", rerr)
		return
	}
	err := request.QueryFields("Name")
	if !assert.NoError(t, err) {
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalBy("Name")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, len(result.Values))
	if !assert.Equal(t, 2, len(result.Data)) {
		return
	}
	result.DumpData()
}
