package adabas

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var vehicleMapName = mapVehicles + "Go"

func prepareStoreAndHold(t *testing.T, c chan bool) {
	log.Infof("Create connection...")
	connection, err := NewConnection("acj;map;config=[" + adabasModDBIDs + ",250]")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	defer connection.Close()

	storeRequest16, err := connection.CreateMapStoreRequest(massLoadSystransStore)
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	recErr := storeRequest16.StoreFields("PERSONNEL-ID,FULL-NAME")
	if !assert.NoError(t, recErr) {
		c <- false
		return
	}
	err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_0")
	if err != nil {
		c <- false
		return
	}
	storeRequest19, cErr := connection.CreateMapStoreRequest(vehicleMapName)
	if !assert.NoError(t, cErr) {
		c <- false
		return
	}
	recErr = storeRequest19.StoreFields("REG-NUM,CAR-DETAILS")
	if !assert.NoError(t, recErr) {
		c <- false
		return
	}
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_0")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	for i := 1; i < 10; i++ {
		x := strconv.Itoa(i)
		err = addEmployeeRecord(t, storeRequest16, multipleTransactionRefName+"_"+x)
		if !assert.NoError(t, err) {
			c <- false
			return
		}

	}
	err = addVehiclesRecord(t, storeRequest19, multipleTransactionRefName2+"_1")
	if !assert.NoError(t, err) {
		c <- false
		return
	}
	c <- true
	time.Sleep(10 * time.Second)
	fmt.Println("End transaction")
	connection.EndTransaction()
	c <- true

}

func TestConnectionTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	f := initTestLogWithFile(t, "connection_transaction.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())

	cErr := clearFile(16)
	if !assert.NoError(t, cErr) {
		return
	}
	cErr = clearFile(19)
	if !assert.NoError(t, cErr) {
		return
	}

	log.Infof("Prepare create test map")
	dataRepository := &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 16}
	perr := prepareCreateTestMap(massLoadSystransStore, massLoadSystrans, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}
	dataRepository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 19}
	perr = prepareCreateTestMap(vehicleMapName, vehicleSystransStore, dataRepository)
	if !assert.NoError(t, perr) {
		return
	}

	c := make(chan bool)
	go prepareStoreAndHold(t, c)
	x := <-c
	if !x && t.Failed() {
		return
	}
	x = <-c

	fmt.Println("Check stored data", x)
	log.Infof("Check stored data")
	checkStoreByFile(t, adabasModDBIDs, 16, multipleTransactionRefName)
	checkStoreByFile(t, adabasModDBIDs, 19, multipleTransactionRefName2)

}
