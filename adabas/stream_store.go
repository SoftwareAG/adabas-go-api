package adabas

import (
	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// ReadLOBRecord read lob records in an stream, repeated call will read next segment of LOB
func (request *StoreRequest) UpdateLOBRecord(isn adatypes.Isn, field string, offset uint64, data []byte) (err error) {
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Store LOB record initiated ...")
	}
	err = request.Open()
	if err != nil {
		return
	}
	err = request.StoreFields(field)
	if err != nil {
		adatypes.Central.Log.Debugf("Store fields error ...%#v", err)
		return err
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("LOB Definition generated ...BlockSize=%d", len(data))
	}
	var record *Record
	record, err = request.CreateRecord()
	if err != nil {
		return
	}
	record.Isn = isn
	err = record.SetPartialValue(field, uint32(offset+1), data)
	if err != nil {
		adatypes.Central.Log.Debugf("Set partial value error ...%#v", err)
		return err
	}
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Update LOB with ...%#v", field)
	}

	adabasRequest, prepareErr := request.prepareRequest(false)
	if prepareErr != nil {
		return prepareErr
	}
	err = request.update(adabasRequest, record)
	if adatypes.Central.IsDebugLevel() {
		adatypes.Central.Log.Debugf("Error reading %v", err)
	}

	return err
}
