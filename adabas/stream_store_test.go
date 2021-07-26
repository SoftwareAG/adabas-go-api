package adabas

import (
	"fmt"
	"os"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

type lobTest struct {
	Index uint64 `adabas:":isn"`
	Name  string `adabas:"::BB"`
}

func TestStreamStore(t *testing.T) {
	initTestLogWithFile(t, "stream_store.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs)
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	storeRequest, serr := connection.CreateStoreRequest(202)
	if !assert.NoError(t, serr) {
		return
	}
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	}
	p = p + "/../files/img/106-0687_IMG.JPG"
	f, err := os.Open(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if !assert.NoError(t, err) {
		return
	}
	data := make([]byte, fi.Size())
	var n int
	n, err = f.Read(data)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Printf("Number of bytes read: %d/%d\n", n, len(data))
	lobEntry := &lobTest{Name: "1234Test.JPG"}
	err = storeRequest.StoreData(lobEntry)
	assert.NoError(t, err)
	fmt.Println("ISN:", lobEntry.Index)
	err = storeRequest.EndTransaction()
	assert.NoError(t, err)

	from := uint64(0)
	blocksize := uint64(8096)
	for {
		err = storeRequest.UpdateLOBRecord(adatypes.Isn(lobEntry.Index), "DC", from, data[from:int(from+blocksize)])
		if !assert.NoError(t, err) {
			return
		}
		from += blocksize
		if int(from) > len(data) {
			break
		}
		if len(data) < int(from+blocksize) {
			blocksize = uint64(len(data)) % blocksize
		}
	}
}
