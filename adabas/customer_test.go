package adabas

import (
	"fmt"
	"testing"
)

func TestCustomer_ProgressLentsAPI(t *testing.T) {
	dburl, _ := NewURL("12(adatcp://192.168.0.27:60001)")
	maprepo := NewMapRepositoryWithURL(DatabaseURL{URL: *dburl, Fnr: 4})
	monmap, _ := LoadJSONMap("daniel_map.json")

	connection, err := NewConnection("acj;map;config=[1(adatcp://localhost:60001),4]")
	if err != nil {
		fmt.Printf("NewConnection() error=%v\n", err)
		return
	}
	defer connection.Close()

	err = connection.Open()
	fmt.Println("MAP LOADED FROM JSON FILE")
	fmt.Printf("%v\n", monmap[0].String())

	if err != nil {
		fmt.Printf("Open() error=%v\n", err)
		return
	}

	monmap[0].Repository = &maprepo.DatabaseURL

	fmt.Println("ADDING MAP TO GLOBAL REPO")
	fmt.Println("\tADDING FNR 4 AS GLOBAL REPO")

	addmaperr := AddGlobalMapRepositoryReference("1(adatcp://localhost:60001),4")
	if addmaperr != nil {
		fmt.Printf("global map add error:%v\n", addmaperr)
	}

	fmt.Println("\tSTORE MAP TO REPO")
	storeerr := monmap[0].Store()
	if storeerr != nil {
		fmt.Printf("store map error=%v\n", storeerr)
	}

	fmt.Println("ADDING MAP TO CACHE")

	maprepo.AddMapToCache("DempMap", monmap[0])
	adadb, dberr := NewAdabas("12(adatcp://localhost:60001)")

	if dberr != nil {
		fmt.Printf("db error:%v\n", dberr)
	}

	fmt.Println("READING ALL GLOBAL MAPS")

	glbmaps, glberr := AllGlobalMaps(adadb)
	if glberr != nil {
		fmt.Printf("glb maps error: %v\n", glberr)
	}

	fmt.Printf("%v\n", glbmaps)

	fmt.Println("DUMPING GLOBAL MAP REPO")
	DumpGlobalMapRepositories()

}
