package adabas

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogicalWithQueryFieldsScan1(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request := NewReadRequestAdabas(adabas, 11)
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

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 299, id)
	assert.Equal(t, "11300317", x)
	assert.Equal(t, "HORST", y)
	assert.Equal(t, "WERNER", z)

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 359, id)
	assert.Equal(t, "11600314", x)
	assert.Equal(t, "BEATE", y)
	assert.Equal(t, "BIRGIT", z)

	err = request.Scan(&id, &x, &y, &z)
	assert.NoError(t, err)

	fmt.Println("Scan result id=", id, "AA=", x, "AC=", y, "AD=", z)
	assert.Equal(t, 365, id)
	assert.Equal(t, "11500307", x)
	assert.Equal(t, "MARA", y)
	assert.Equal(t, "", z)

	err = request.Scan(&id, &x, &y, &z)
	assert.Error(t, err)
	assert.Equal(t, "EOF", err.Error())

}

func TestRequestLogicalWithQueryFieldsScan2(t *testing.T) {
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request := NewReadRequestAdabas(adabas, 9)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request := NewReadRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request := NewReadRequestAdabas(adabas, 11)
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
	f := initTestLogWithFile(t, "request.log")
	defer f.Close()

	log.Infof("TEST: %s", t.Name())
	adabas, err := NewAdabas(adabasStatDBID)
	if !assert.NoError(t, err) {
		fmt.Println("Error new adabas", err)
		return
	}
	request := NewReadRequestAdabas(adabas, 11)
	if !assert.NotNil(t, request) {
		fmt.Println("Error request nil")
		return
	}
	defer request.Close()
	err = request.QueryFields("AA,#ISNQUANTITY")
	if !assert.NoError(t, err) {
		fmt.Println("Error query fields", err)
		return
	}

	_, err = request.HistogramByCursoring("AA")
	assert.NoError(t, err)
	var id int
	err = nil
	i := 0
	for err == nil {
		var x string
		err = request.Scan(&x, &id)
		if !assert.NoError(t, err) {
			return
		}
		i++
		fmt.Printf("%d. Scan result id=%d AA=%s\n", i, id, x)

	}

}
