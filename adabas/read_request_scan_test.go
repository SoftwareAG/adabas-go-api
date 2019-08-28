/*
* Copyright © 2018-2019 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */

package adabas

import (
	"fmt"
	"testing"

	"github.com/SoftwareAG/adabas-go-api/adatypes"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogicalWithQueryFieldsScan1(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("#ISN,AA,AC,AD")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.ReadLogicalWithCursoring("AE=MUELLER")
	assert.NoError(t, err)
	var id int
	var x, y, z string
	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 255, id)
	assert.Equal(t, "11100308", x)
	assert.Equal(t, "DIETER", y)
	assert.Equal(t, "PETER", z)
	x = ""
	y = ""
	z = ""

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 299, id)
	assert.Equal(t, "11300317", x)
	assert.Equal(t, "HORST", y)
	assert.Equal(t, "WERNER", z)
	x = ""
	y = ""
	z = ""

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 359, id)
	assert.Equal(t, "11600314", x)
	assert.Equal(t, "BEATE", y)
	assert.Equal(t, "BIRGIT", z)
	x = ""
	y = ""
	z = ""

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 365, id)
	assert.Equal(t, "11500307", x)
	assert.Equal(t, "MARA", y)
	assert.Equal(t, "", z)

	err = request.Scan(&id, &x, &y, &z)
	assert.Error(t, err)
	assert.Equal(t, "ADG0000130: End of cursor", err.Error())

}

func TestRequestLogicalWithQueryFieldsScan2(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 9)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("#ISN,AA,BC,KA,JA,MA,EA")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.ReadLogicalWithCursoring("BC=Müller")
	assert.NoError(t, err)
	var id int
	var x, y, z string
	var m float32
	m = 5.0
	err = request.Scan(&id, &x, &y, &z, nil, &m)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "BC=", y, "KA=", z, "MA=", m)
	assert.Equal(t, 254, id)
	assert.Equal(t, "11100308", x)
	assert.Equal(t, "Müller", y)
	assert.Equal(t, "Auszeichner", z)
	assert.Equal(t, float32(0.0), m)

}

func TestRequestLogicalWithQueryFieldsScan3(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("#ISN,AA,AI")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.ReadLogicalWithCursoring("AE=HOLMES")
	assert.NoError(t, err)
	var id int
	var x, y string
	err = request.Scan(&id, &x, &y)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AI=", y)
	assert.Equal(t, 886, id)
	assert.Equal(t, "30000651", x)
	assert.Equal(t, "14 MAIN STREET,MELBOURNE,DERBYSHIRE", y)

}

func TestRequestLogicalWithQueryFieldsScan4(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("#ISN,AA,AE,AZ,AR,AT")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.ReadLogicalWithCursoring("AE=ADAM")
	assert.NoError(t, err)
	var id int
	var x, y, z string
	var ar []string
	var at []string
	err = request.Scan(&id, &x, &y, &z, &ar, &at)
	if !assert.NoError(t, err) {
		return
	}

	fmt.Println("Scan result id=", id, "AA=", x, "AE=", y, "AZ=", z, "AR=", ar, "AT=", at)
	assert.Equal(t, 1, id)
	assert.Equal(t, "50005800", x)
	assert.Equal(t, "ADAM", y)
	assert.Equal(t, "FRE,ENG", z)

}

func TestHistogramByQueryFieldsScan(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("AE,#ISNQUANTITY")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.HistogramByCursoring("AE")
	assert.NoError(t, err)
	var id int
	err = nil
	i := 0
	var checked [4]bool
	for err == nil {
		var x string
		err = request.Scan(&x, &id)
		if err == nil {
			switch x {
			case "SMITH":
				if !assert.False(t, checked[0]) {
					return
				}
				checked[0] = true
				assert.Equal(t, 19, id)
				fmt.Printf("%d. Scan result quantity=%d AA=%s\n", i, id, x)
			case "WOOD":
				if !assert.False(t, checked[1]) {
					return
				}
				checked[1] = true
				assert.Equal(t, 2, id)
				fmt.Printf("%d. Scan result quantity=%d AA=%s\n", i, id, x)
			case "JONES":
				if !assert.False(t, checked[2]) {
					return
				}
				checked[2] = true
				assert.Equal(t, 9, id)
				fmt.Printf("%d. Scan result quantity=%d AA=%s\n", i, id, x)
			case "FERNANDEZ":
				if !assert.False(t, checked[3]) {
					return
				}
				checked[3] = true
				assert.Equal(t, 9, id)
				fmt.Printf("%d. Scan result quantity=%d AA=%s\n", i, id, x)
			default:
			}
			if !assert.True(t, id > 0 && i < 805, fmt.Sprintf("Quantity is %d counter=%d", id, i)) {
				return
			}
			i++
		}
	}
	assert.Equal(t, 804, i)

}

func TestHistogramWithQueryFieldsScan(t *testing.T) {
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request, _ := NewReadRequest(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("AE,#ISNQUANTITY")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.HistogramWithCursoring("AE=[A:B]")
	assert.NoError(t, err)
	var id int
	err = nil
	i := 0
	for err == nil {
		var x string
		err = request.Scan(&x, &id)
		if err == nil {
			switch x {
			case "ADKINSON":
				assert.Equal(t, 8, id)
				fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)
			case "ALEMAN":
				assert.Equal(t, 1, id)
				fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)
			case "ALEXANDER":
				assert.Equal(t, 5, id)
				fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)
			case "ANDERSEN":
				assert.Equal(t, 3, id)
				fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)
			default:
			}
			assert.True(t, id > 0, fmt.Sprintf("Quantity is %d", id))
			//			fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)
			i++
		}
	}
	assert.Equal(t, 20, i)

}

func TestConnectionSimpleScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs + ";auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("AA,AC,AD,AE")
	assert.NoError(t, err)

	adatypes.Central.Log.Debugf("Test Search with ...")
	result, rerr := readRequest.ReadLogicalWith("AE=ADAM")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Found entries:", result.NrRecords())
	assert.Equal(t, 1, result.NrRecords())
	for _, record := range result.Values {
		var aa, ac, ad, ae string
		// Read given AA(alpha) and all entries of group AB to string variables
		record.Scan(&aa, &ac, &ad, &ae)
		fmt.Println(aa, ac, ad, ae)
		assert.Equal(t, "50005800", aa)
		assert.Equal(t, "SIMONE", ac)
		assert.Equal(t, "", ad)
		assert.Equal(t, "ADAM", ae)
	}

}

func TestConnectionSimpleScanGroupReference(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs + ";auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("AA,AB")
	assert.NoError(t, err)

	adatypes.Central.Log.Debugf("Test Search with ...")
	result, rerr := readRequest.ReadLogicalWith("AE=ADAM")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Found entries:", result.NrRecords())
	assert.Equal(t, 1, result.NrRecords())
	for _, record := range result.Values {
		var aa, ac, ad, ae string
		// Read given AA(alpha) and all entries of group AB to string variables
		record.Scan(&aa, &ac, &ae, &ad)
		fmt.Println(aa, ac, ad, ae)
		assert.Equal(t, "50005800", aa)
		assert.Equal(t, "SIMONE", ac)
		assert.Equal(t, "", ad)
		assert.Equal(t, "ADAM", ae)
	}

}

func TestConnectionSimpleScanGroupAndField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	initTestLogWithFile(t, "request.log")

	adatypes.Central.Log.Infof("TEST: %s", t.Name())
	connection, err := NewConnection("acj;target=" + adabasModDBIDs + ";auth=DESC,user=TCMapPoin,id=4,host=UNKNOWN")
	if !assert.NoError(t, err) {
		return
	}
	defer connection.Close()

	connection.Open()
	readRequest, rErr := connection.CreateFileReadRequest(11)
	assert.NoError(t, rErr)
	err = readRequest.QueryFields("AA,AB,AO")
	assert.NoError(t, err)

	adatypes.Central.Log.Debugf("Test Search with ...")
	result, rerr := readRequest.ReadLogicalWith("AE=ADAM")
	if !assert.NoError(t, rerr) {
		return
	}
	fmt.Println("Found entries:", result.NrRecords())
	assert.Equal(t, 1, result.NrRecords())
	for _, record := range result.Values {
		var aa, ac, ad, ae, ao string
		// Read given AA(alpha) and all entries of group AB to string variables
		record.Scan(&aa, &ac, &ae, &ad, &ao)
		fmt.Println(aa, ac, ad, ae, ao)
		assert.Equal(t, "50005800", aa)
		assert.Equal(t, "SIMONE", ac)
		assert.Equal(t, "", ad)
		assert.Equal(t, "ADAM", ae)
		assert.Equal(t, "VENT59", ao)
	}

}
