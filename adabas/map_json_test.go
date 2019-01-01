package adabas

import (
	"fmt"
	"os"
	"testing"
)

func TestMapImportJson(t *testing.T) {
	f := initTestLogWithFile(t, "mapjson.log")
	defer f.Close()

	adabas := NewAdabas(23)
	defer adabas.Close()
	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + "/" + "Maps.json"
	fmt.Println("Loading ...." + name)
	file, err := os.Open(name)
	if err != nil {
		return
	}
	defer file.Close()

	maps, err := ParseJSONFileForFields(file)
	fmt.Println("Number of maps", len(maps), err)
	for _, m := range maps {
		fmt.Println("MAP", m.Name)
		fmt.Println(" ", m.Data.URL.String(), m.Data.Fnr)
		for _, f := range m.Fields {
			fmt.Println("   ", f.LongName, f.ShortName, f.Length, f.FormatType, f.ContentType)
		}

	}

}
