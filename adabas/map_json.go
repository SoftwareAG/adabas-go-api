package adabas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
)

// MapFile parse map JSON import/export files
type MapFile struct {
	Maps []*Map `json:"Maps"`
}

// ParseJSONFileForFields Parse JSON map export file
func ParseJSONFileForFields(file *os.File) (mapList []*Map, err error) {

	byteValue, _ := ioutil.ReadAll(file)

	var mapFile MapFile
	err = json.Unmarshal([]byte(byteValue), &mapFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	adatypes.Central.Log.Debugf("Number map entries %d", len(mapFile.Maps))
	mapList = mapFile.Maps
	return
}
