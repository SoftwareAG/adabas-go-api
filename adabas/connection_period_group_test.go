package adabas

import (
	"fmt"
	//	log "github.com/sirupsen/logrus"
	//	"github.com/stretchr/testify/assert"
	//	"testing"
)

func ExampleConnection_PeriodGroup() {
	f, _ := initLogWithFile("connection.log")
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error open connection", cerr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100303:11100304]")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = result.DumpValues()
	if err != nil {
		fmt.Println("Error dump values", err)
	}

	// Output: Dump all result values
	// Record Isn: 0252
	//   PERSONNEL-ID = > 11100303 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > KRISTINA             <
	//    NAME = > FALTER               <
	//    MIDDLE-I = > M <
	//   MAR-STAT = > M <
	//   SEX = > F <
	//   BIRTH = > 716428 <
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 3 ]
	//     ADDRESS-LINE[01] = > C/O CLAASEN          <
	//     ADDRESS-LINE[02] = > WIESENGRUND 10       <
	//     ADDRESS-LINE[03] = > 6100 DARMSTADT       <
	//    CITY = > DARMSTADT            <
	//    POST-CODE = > 6100       <
	//    COUNTRY = > D   <
	//   TELEPHONE = [ 1 ]
	//    AREA-CODE = > 06151  <
	//    PHONE = > 453897          <
	//   DEPT = > FINA21 <
	//   JOB-TITLE = > TYPISTIN                  <
	//   INCOME = [ 3 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 21846 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 1717 <
	//     BONUS[01,02] = > 3000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 21025 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1538 <
	//    CURR-CODE[03] = > EUR <
	//    SALARY[03] = > 20307 <
	//    BONUS[03] = [ 1 ]
	//     BONUS[03,01] = > 1282 <
	//   LEAVE-DATA = [ 1 ]
	//    LEAVE-DUE = > 30 <
	//    LEAVE-TAKEN = > 3 <
	//   LEAVE-BOOKED = [ 1 ]
	//    LEAVE-START[01] = > 19980520 <
	//    LEAVE-END[01] = > 19980523 <
	//   LANG = [ 1 ]
	//    LANG[01] = > GER <
	//   PH = >  <
	//   H1 = > 3003 <
	//   S1 = > FINA <
	//   S2 = > FINA21FALTER               <
	//   S3 = >  <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   FULL-NAME = [ 1 ]
	//    FIRST-NAME = > WOLFGANG             <
	//    NAME = > SCHMIDT              <
	//    MIDDLE-I = > J <
	//   MAR-STAT = > M <
	//   SEX = > M <
	//   BIRTH = > 709788 <
	//   FULL-ADDRESS = [ 1 ]
	//    ADDRESS-LINE = [ 3 ]
	//     ADDRESS-LINE[01] = > POSTFACH 67          <
	//     ADDRESS-LINE[02] = > MANDELA-WEG 8        <
	//     ADDRESS-LINE[03] = > 6000 FRANKFURT       <
	//    CITY = > FRANKFURT            <
	//    POST-CODE = > 6000       <
	//    COUNTRY = > D   <
	//   TELEPHONE = [ 1 ]
	//    AREA-CODE = > 069    <
	//    PHONE = > 549987          <
	//   DEPT = > FINA21 <
	//   JOB-TITLE = > SACHBEARBEITER            <
	//   INCOME = [ 2 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 25230 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 2256 <
	//     BONUS[01,02] = > 2000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 24102 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1948 <
	//   LEAVE-DATA = [ 1 ]
	//    LEAVE-DUE = > 30 <
	//    LEAVE-TAKEN = > 0 <
	//   LEAVE-BOOKED = [ 0 ]
	//   LANG = [ 2 ]
	//    LANG[01] = > GER <
	//    LANG[02] = > ENG <
	//   PH = >  <
	//   H1 = > 3000 <
	//   S1 = > FINA <
	//   S2 = > FINA21SCHMIDT              <
	//   S3 = >  <

}

func ExampleConnection_PeriodGroupPart() {
	f, _ := initLogWithFile("connection.log")
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error open connection", cerr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	request.QueryFields("PERSONNEL-ID,INCOME")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100303:11100304]")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = result.DumpValues()
	if err != nil {
		fmt.Println("Error dump values", err)
	}

	// Output: Dump all result values
	// Record Isn: 0252
	//   PERSONNEL-ID = > 11100303 <
	//   INCOME = [ 3 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 21846 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 1717 <
	//     BONUS[01,02] = > 3000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 21025 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1538 <
	//    CURR-CODE[03] = > EUR <
	//    SALARY[03] = > 20307 <
	//    BONUS[03] = [ 1 ]
	//     BONUS[03,01] = > 1282 <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   INCOME = [ 2 ]
	//    CURR-CODE[01] = > EUR <
	//    SALARY[01] = > 25230 <
	//    BONUS[01] = [ 2 ]
	//     BONUS[01,01] = > 2256 <
	//     BONUS[01,02] = > 2000 <
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 24102 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1948 <

}

func ExampleConnection_PeriodGroupLastEntry() {
	f, _ := initLogWithFile("connection.log")
	defer f.Close()

	connection, cerr := NewConnection("acj;map;config=[" + adabasStatDBIDs + ",4]")
	if cerr != nil {
		fmt.Println("Error new connection", cerr)
		return
	}
	defer connection.Close()
	openErr := connection.Open()
	if openErr != nil {
		fmt.Println("Error open connection", cerr)
		return
	}

	request, err := connection.CreateMapReadRequest("EMPLOYEES-NAT-DDM")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	request.QueryFields("PERSONNEL-ID,INCOME[N]")
	request.Limit = 0
	var result *Response
	result, err = request.ReadLogicalWith("PERSONNEL-ID=[11100303:11100304]")
	if err != nil {
		fmt.Println("Error create request", err)
		return
	}
	err = result.DumpValues()
	if err != nil {
		fmt.Println("Error dump values", err)
	}

	// Output: Dump all result values
	// Record Isn: 0252
	//   PERSONNEL-ID = > 11100303 <
	//   INCOME = [ 3 ]
	//    CURR-CODE[03] = > EUR <
	//    SALARY[03] = > 20307 <
	//    BONUS[03] = [ 1 ]
	//     BONUS[03,01] = > 1282 <
	// Record Isn: 0253
	//   PERSONNEL-ID = > 11100304 <
	//   INCOME = [ 2 ]
	//    CURR-CODE[02] = > EUR <
	//    SALARY[02] = > 24102 <
	//    BONUS[02] = [ 1 ]
	//     BONUS[02,01] = > 1948 <

}
