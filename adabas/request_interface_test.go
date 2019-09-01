package adabas

import (
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

// type Employees struct {
// 	FirstName string
// 	Name      string
// 	Sec       string
// }

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

	employees := make([]*Employees, 0)
	employees = append(employees, &Employees{ID: "ID", Birth: 123, Name: "Name", FirstName: "First name"})
	employees = append(employees, &Employees{ID: "ID2", Birth: 234, Name: "Name2", FirstName: "First name2"})
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

func TestReadInterface(t *testing.T) {
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

	result, err := request.ReadLogicalWith("ID>'ID'")
	fmt.Println("Read done ...")
	if !assert.NoError(t, err) {
		return
	}
	assert.Nil(t, result.Values)
	assert.NotNil(t, result.Data)
	if assert.NotNil(t, result) {
		result.DumpValues()
	}

}
