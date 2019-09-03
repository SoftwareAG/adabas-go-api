package adabas

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type EmployeesSalary struct {
	ID         string `adabas:"Id"`
	FirstName  string
	LastName   string
	Department string
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

func TestStoreInterface(t *testing.T) {
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

	result, err := request.ReadLogicalWith("ID>'ID'")
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
		assert.Len(t, result.Data, 3)
		if assert.IsType(t, (*Employees)(nil), result.Data[0]) {
			e := result.Data[0].(*Employees)
			assert.Equal(t, "ID2", strings.Trim(e.ID, " "))
			assert.Equal(t, int64(234), e.Birth)
			assert.Equal(t, "Name2", strings.Trim(e.Name, " "))
			e = result.Data[1].(*Employees)
			assert.Equal(t, "ID3", strings.Trim(e.ID, " "))
			assert.Equal(t, "Name3", strings.Trim(e.Name, " "))
			e = result.Data[2].(*Employees)
			assert.Equal(t, "ID4", strings.Trim(e.ID, " "))
			assert.Equal(t, "Name4", strings.Trim(e.Name, " "))

		}
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
	result, err := request.ReadLogicalWithInterface("ID>'ID'", receiveInterface, &i)
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 3, i)
	assert.Nil(t, result.Values)
	assert.Nil(t, result.Data)
}
