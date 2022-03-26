/*
* Copyright © 2019-2022 Software AG, Darmstadt, Germany and/or its licensors
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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleParseJSONFileForFields() {
	lerr := initLogWithFile("mapjson.log")
	if lerr != nil {
		fmt.Println("Init log error", lerr)
		return
	}

	p := os.Getenv("TESTFILES")
	if p == "" {
		p = "."
	}
	name := p + string(os.PathSeparator) + "Maps.json"
	fmt.Println("Loading ....Maps.json")
	file, err := os.Open(name)
	if err != nil {
		return
	}
	defer file.Close()

	maps, err := ParseJSONFileForFields(file)
	if err != nil {
		fmt.Println("Error parsing file", err)
		return
	}
	fmt.Println("Number of maps", len(maps))
	for _, m := range maps {
		fmt.Println("MAP", m.Name)
		fmt.Printf("  %s %d\n", m.Data.URL.String(), m.Data.Fnr)
		for _, f := range m.Fields {
			fmt.Printf("   ln=%s sn=%s len=%d format=%s content=%s\n", f.LongName, f.ShortName, f.Length, f.FormatType, f.ContentType)
		}

	}

	// Output: Loading ....Maps.json
	// Number of maps 17
	// MAP NEW_EMPLOYEES
	//   23 9
	//    ln=personnel-data sn=A0 len=8 format=  content=
	//    ln=personnel-id sn=AA len=8 format=A content=
	//    ln=id-data sn=AB len=12 format=  content=
	//    ln=personnel-no sn=AC len=4 format=I content=
	//    ln=id-card sn=AD len=8 format=B content=
	//    ln=signature sn=AE len=0 format=A content=
	//    ln=full-name sn=B0 len=130 format=  content=
	//    ln=first-name sn=BA len=40 format=U content=charset=UTF-8
	//    ln=middle-name sn=BB len=40 format=U content=charset=UTF-8
	//    ln=name sn=BC len=50 format=U content=charset=UTF-8
	//    ln=mar-stat sn=CA len=1 format=A content=
	//    ln=sex sn=DA len=1 format=A content=
	//    ln=birth sn=EA len=4 format=D content=
	//    ln=private-address sn=F0 len=113 format=  content=
	//    ln=p-address-line sn=FA len=60 format=U content=charset=UTF-8
	//    ln=p-address-line sn=FA len=60 format=U content=charset=UTF-8
	//    ln=p-city sn=FB len=40 format=U content=charset=UTF-8
	//    ln=p-post-code sn=FC len=10 format=A content=
	//    ln=p-country sn=FD len=3 format=A content=
	//    ln=p-phone-email sn=F1 len=131 format=  content=
	//    ln=p-area-code sn=FE len=6 format=A content=
	//    ln=private-phone sn=FF len=15 format=A content=
	//    ln=private-fax sn=FG len=15 format=A content=
	//    ln=private-mobile sn=FH len=15 format=A content=
	//    ln=private-email sn=FI len=80 format=A content=
	//    ln=private-email sn=FI len=80 format=A content=
	//    ln=business-address sn=I0 len=98 format=  content=
	//    ln=b-address-line sn=IA len=40 format=U content=charset=UTF-8
	//    ln=b-address-line sn=IA len=40 format=U content=charset=UTF-8
	//    ln=b-city sn=IB len=40 format=U content=charset=UTF-8
	//    ln=b-post-code sn=IC len=10 format=A content=
	//    ln=b-country sn=ID len=3 format=A content=
	//    ln=room_number sn=IE len=5 format=A content=
	//    ln=b-phone-email sn=I1 len=131 format=  content=
	//    ln=b-area-code sn=IF len=6 format=A content=
	//    ln=business-phone sn=IG len=15 format=A content=
	//    ln=business-fax sn=IH len=15 format=A content=
	//    ln=business-mobile sn=II len=15 format=A content=
	//    ln=business-email sn=IJ len=80 format=A content=
	//    ln=business-email sn=IJ len=80 format=A content=
	//    ln=job-title sn=KA len=66 format=U content=charset=UTF-8
	//    ln=income sn=L0 len=15 format=  content=
	//    ln=curr-code sn=LA len=3 format=A content=
	//    ln=salary sn=LB len=6 format=P content=
	//    ln=bonus sn=LC len=6 format=P content=
	//    ln=bonus sn=LC len=6 format=P content=
	//    ln=total-income sn=MA len=4 format=F content=
	//    ln=leave-date sn=N0 len=5 format=  content=
	//    ln=leave-due sn=NA len=2 format=N content=
	//    ln=leave-taken sn=NB len=3 format=N content=
	//    ln=leave-booked sn=O0 len=16 format=  content=
	//    ln=leave-start sn=OA len=8 format=N content=
	//    ln=leave-end sn=OB len=8 format=N content=
	//    ln=language sn=PA len=3 format=A content=
	//    ln=language sn=PA len=3 format=A content=
	//    ln=last-update sn=QA len=7 format=P content=
	//    ln=picture sn=RA len=0 format=A content=
	//    ln=documents sn=S0 len=83 format=  content=
	//    ln=document-description sn=SA len=80 format=U content=charset=UTF-8
	//    ln=document-type sn=SB len=3 format=A content=
	//    ln=document sn=SC len=0 format=A content=
	//    ln=document sn=SC len=0 format=A content=
	//    ln=creation_time sn=TC len=20 format=N content=
	//    ln=Last_Updates sn=TU len=20 format=N content=
	//    ln=Last_Updates sn=TU len=20 format=N content=
	// MAP ADABAS_MAP
	//   23 4
	//    ln=TYPE-GROUP sn=TY len=0 format=  content=
	//    ln=TYPE sn=TA len=1 format=B content=
	//    ln=GENERATED sn=AA len=28 format=  content=
	//    ln=HOST sn=AB len=20 format=A content=
	//    ln=DATE sn=AC len=8 format=B content=
	//    ln=VERSION sn=AD len=2 format=B content=
	//    ln=LOCATION sn=RA len=45 format=  content=
	//    ln=FILE-NR sn=RF len=4 format=B content=
	//    ln=DATA-DATABASE sn=RD len=20 format=A content=
	//    ln=MAPNAME sn=RN len=20 format=A content=
	//    ln=MAP-FLAGS sn=RB len=1 format=B content=
	//    ln=MAP-OPTIONS sn=RO len=20 format=A content=
	//    ln=INDEX sn=DL len=4 format=  content=
	//    ln=INDEX-FILE sn=DF len=4 format=B content=
	//    ln=INDEX-DATABASE sn=DD len=20 format=A content=
	//    ln=MAPPING sn=MA len=68 format=  content=
	//    ln=SHORTNAME sn=MB len=20 format=A content=
	//    ln=TYPE-CONVERSION sn=MC len=4 format=B content=
	//    ln=LONGNAME sn=MD len=20 format=A content=
	//    ln=LENGTH sn=ML len=4 format=B content=
	//    ln=CONTENTTYPE sn=MT len=20 format=A content=
	//    ln=FORMATTYPE sn=MY len=2 format=A content=
	//    ln=TIMESTAMP sn=ZB len=8 format=B content=
	//    ln=TIMESTAMP sn=ZB len=8 format=B content=
	// MAP EMPLOYEES-NAT-DDM
	//   23 11
	//    ln=PERSONNEL-ID sn=AA len=8 format=A content=
	//    ln=FULL-NAME sn=AB len=21 format=  content=
	//    ln=FIRST-NAME sn=AC len=20 format=A content=
	//    ln=NAME sn=AE len=20 format=A content=
	//    ln=MIDDLE-I sn=AD len=1 format=A content=
	//    ln=MAR-STAT sn=AF len=1 format=A content=
	//    ln=SEX sn=AG len=1 format=A content=
	//    ln=BIRTH sn=AH len=4 format=D content=
	//    ln=FULL-ADDRESS sn=A1 len=50 format=  content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=CITY sn=AJ len=20 format=A content=
	//    ln=ZIP sn=AK len=10 format=A content=
	//    ln=COUNTRY sn=AL len=3 format=A content=
	//    ln=TELEPHONE sn=A2 len=6 format=  content=
	//    ln=AREA-CODE sn=AN len=6 format=A content=
	//    ln=PHONE sn=AM len=15 format=A content=
	//    ln=DEPT sn=AO len=6 format=A content=
	//    ln=JOB-TITLE sn=AP len=25 format=A content=
	//    ln=INCOME sn=AQ len=12 format=  content=
	//    ln=CURR-CODE sn=AR len=3 format=A content=
	//    ln=SALARY sn=AS len=9 format=P content=
	//    ln=BONUS sn=AT len=9 format=P content=
	//    ln=BONUS sn=AT len=9 format=P content=
	//    ln=LEAVE-DATA sn=A3 len=2 format=  content=
	//    ln=LEAVE-DUE sn=AU len=2 format=N content=
	//    ln=LEAVE-TAKEN sn=AV len=2 format=N content=
	//    ln=LEAVE-BOOKED sn=AW len=8 format=  content=
	//    ln=LEAVE-START sn=AX len=8 format=N content=
	//    ln=LEAVE-END sn=AY len=8 format=N content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=PH sn=PH len=0 format=A content=
	//    ln=LEAVE-LEFT sn=H1 len=0 format=A content=
	//    ln=DEPARTMENT sn=S1 len=0 format=A content=
	//    ln=DEPT-PERSON sn=S2 len=0 format=A content=
	//    ln=CURRENCY-SALARY sn=S3 len=0 format=A content=
	// MAP LOBEXAMPLE
	//   23 202
	//    ln=Generated sn=AA len=16 format=  content=
	//    ln=Host sn=AB len=0 format=A content=
	//    ln=Date sn=AC len=8 format=B content=
	//    ln=Version sn=AD len=8 format=B content=
	//    ln=Location sn=BA len=0 format=  content=
	//    ln=Directory sn=BB len=0 format=U content=charset=UTF-8
	//    ln=Filename sn=BC len=0 format=U content=charset=UTF-8
	//    ln=absoluteFilename sn=BD len=0 format=U content=charset=UTF-8
	//    ln=Type sn=CA len=4 format=  content=
	//    ln=Size sn=CB len=4 format=B content=
	//    ln=MimeType sn=CC len=0 format=U content=charset=UTF-8
	//    ln=Data sn=DA len=0 format=  content=
	//    ln=Thumbnail sn=DB len=0 format=A content=
	//    ln=Picture sn=DC len=0 format=A content=
	//    ln=Checksum sn=EA len=80 format=  content=
	//    ln=ThumbnailSHAchecksum sn=EB len=40 format=A content=
	//    ln=PictureSHAchecksum sn=EC len=40 format=A content=
	//    ln=EXIFinformation sn=FA len=0 format=  content=
	//    ln=Model sn=FB len=0 format=U content=charset=UTF-8
	//    ln=Orientation sn=FC len=0 format=U content=charset=UTF-8
	//    ln=DateExif sn=FD len=0 format=U content=charset=UTF-8
	//    ln=DateOriginal sn=FE len=0 format=U content=charset=UTF-8
	//    ln=ExposureTime sn=FF len=0 format=U content=charset=UTF-8
	//    ln=F-Number sn=FG len=0 format=U content=charset=UTF-8
	//    ln=Width sn=FH len=0 format=U content=charset=UTF-8
	//    ln=Height sn=FI len=0 format=U content=charset=UTF-8
	// MAP VEHICLES
	//   23 12
	//    ln=REG-NUM sn=AA len=15 format=A content=
	//    ln=CHASSIS-NUM sn=AB len=4 format=B content=
	//    ln=PERSONNEL-ID sn=AC len=8 format=A content=
	//    ln=CAR-DETAILS sn=CD len=40 format=  content=
	//    ln=MAKE sn=AD len=20 format=A content=
	//    ln=MODEL sn=AE len=20 format=A content=
	//    ln=COLOR sn=AF len=10 format=A content=
	//    ln=YEAR sn=AG len=4 format=N content=
	//    ln=CLASS sn=AH len=1 format=A content=
	//    ln=LEASE-PUR sn=AI len=1 format=A content=
	//    ln=DATE-ACQ sn=AJ len=8 format=N content=
	//    ln=CURR-CODE sn=AL len=3 format=A content=
	//    ln=MAINT-COST sn=AM len=7 format=P content=
	//    ln=MAINT-COST sn=AM len=7 format=P content=
	//    ln=MODEL-YEAR-MAKE sn=AO len=0 format=A content=
	// MAP EmployeeMap
	//   23 11
	//    ln=Id sn=AA len=-1 format=A content=
	//    ln=Name sn=AB len=-1 format=  content=
	//    ln=FirstName sn=AC len=-1 format=A content=
	//    ln=LastName sn=AE len=-1 format=A content=
	//    ln=City sn=AJ len=-1 format=A content=
	//    ln=Department sn=AO len=-1 format=A content=
	//    ln=JobTitle sn=AP len=-1 format=A content=
	//    ln=Income sn=AQ len=-1 format=  content=
	//    ln=Salary sn=AS len=-1 format=A content=
	//    ln=Bonus sn=AT len=-1 format=A content=
	//    ln=Bonus sn=AT len=-1 format=A content=
	// MAP VehicleMap
	//   23 12
	//    ln=Vendor sn=AD len=-1 format=A content=
	//    ln=Model sn=AE len=-1 format=A content=
	//    ln=Color sn=AF len=-1 format=A content=
	// MAP LOB_MAP
	//   23 202
	//    ln=Location sn=BA len=-1 format=  content=
	//    ln=Filename sn=BC len=-1 format=A content=
	//    ln=Type sn=CA len=-1 format=  content=
	//    ln=MimeType sn=CC len=-1 format=A content=
	//    ln=Data sn=DA len=-1 format=  content=
	//    ln=Picture sn=DC len=-1 format=A content=
	//    ln=Checksum sn=EA len=-1 format=  content=
	//    ln=PictureSHAchecksum sn=EC len=-1 format=A content=
	// MAP PictureStore
	//   23 280
	//    ln=Generated sn=AA len=20 format=  content=
	//    ln=Host sn=AB len=0 format=A content=
	//    ln=Date sn=AC len=8 format=B content=
	//    ln=Version sn=AD len=12 format=A content=
	//    ln=Location sn=BA len=0 format=  content=
	//    ln=Directory sn=BB len=0 format=U content=charset=UTF-8
	//    ln=Filename sn=BC len=0 format=U content=charset=UTF-8
	//    ln=absoluteFilename sn=BD len=0 format=U content=charset=UTF-8
	//    ln=Information sn=IN len=0 format=  content=
	//    ln=Description sn=ID len=0 format=U content=charset=UTF-8
	//    ln=Title sn=IT len=0 format=U content=charset=UTF-8
	//    ln=Type sn=CA len=4 format=  content=
	//    ln=Size sn=CB len=4 format=B content=
	//    ln=MimeType sn=CC len=0 format=U content=charset=UTF-8
	//    ln=Data sn=DA len=0 format=  content=
	//    ln=Thumbnail sn=DB len=0 format=A content=
	//    ln=Picture sn=DC len=0 format=A content=
	//    ln=Checksum sn=EA len=80 format=  content=
	//    ln=ThumbnailSHAchecksum sn=EB len=40 format=A content=
	//    ln=PictureSHAchecksum sn=EC len=40 format=A content=
	//    ln=EXIFinformation sn=FA len=24 format=  content=
	//    ln=Model sn=FB len=0 format=U content=charset=UTF-8
	//    ln=Orientation sn=FC len=0 format=U content=charset=UTF-8
	//    ln=DateExif sn=FD len=8 format=P content=
	//    ln=DateOriginal sn=FE len=8 format=P content=
	//    ln=ExposureTime sn=FF len=0 format=U content=charset=UTF-8
	//    ln=F-Number sn=FG len=0 format=U content=charset=UTF-8
	//    ln=Width sn=FH len=4 format=B content=
	//    ln=Height sn=FI len=4 format=B content=
	//    ln=Tag sn=TA len=0 format=  content=
	//    ln=TagName sn=TN len=0 format=U content=charset=UTF-8
	//    ln=TagValue sn=TV len=0 format=U content=charset=UTF-8
	//    ln=Album sn=AL len=0 format=A content=
	//    ln=Album sn=AL len=0 format=A content=
	// MAP LOBSTORE
	//   23 160
	//    ln=Generated sn=AA len=16 format=  content=
	//    ln=Host sn=AB len=0 format=A content=
	//    ln=Date sn=AC len=8 format=B content=
	//    ln=Version sn=AD len=8 format=B content=
	//    ln=Location sn=BA len=0 format=  content=
	//    ln=Directory sn=BB len=0 format=U content=charset=UTF-8
	//    ln=Filename sn=BC len=0 format=U content=charset=UTF-8
	//    ln=absoluteFilename sn=BD len=0 format=U content=charset=UTF-8
	//    ln=Type sn=CA len=4 format=  content=
	//    ln=Size sn=CB len=4 format=B content=
	//    ln=MimeType sn=CC len=0 format=U content=charset=UTF-8
	//    ln=Data sn=DA len=0 format=  content=
	//    ln=Thumbnail sn=DB len=0 format=A content=
	//    ln=Picture sn=DC len=0 format=A content=
	//    ln=Checksum sn=EA len=80 format=  content=
	//    ln=ThumbnailSHAchecksum sn=EB len=40 format=A content=
	//    ln=PictureSHAchecksum sn=EC len=40 format=A content=
	//    ln=EXIFinformation sn=FA len=0 format=  content=
	//    ln=Model sn=FB len=0 format=U content=charset=UTF-8
	//    ln=Orientation sn=FC len=0 format=U content=charset=UTF-8
	//    ln=DateExif sn=FD len=0 format=U content=charset=UTF-8
	//    ln=DateOriginal sn=FE len=0 format=U content=charset=UTF-8
	//    ln=ExposureTime sn=FF len=0 format=U content=charset=UTF-8
	//    ln=F-Number sn=FG len=0 format=U content=charset=UTF-8
	//    ln=Width sn=FH len=0 format=U content=charset=UTF-8
	//    ln=Height sn=FI len=0 format=U content=charset=UTF-8
	// MAP EmployeeX
	//   23 11
	//    ln=PERSONNEL-ID sn=AA len=8 format=A content=
	//    ln=FULL-NAME sn=AB len=0 format=  content=
	//    ln=FIRST-NAME sn=AC len=20 format=A content=
	//    ln=NAME sn=AE len=20 format=A content=
	//    ln=MIDDLE-I sn=AD len=1 format=A content=
	//    ln=MAR-STAT sn=AF len=1 format=A content=
	//    ln=SEX sn=AG len=1 format=A content=
	//    ln=BIRTH sn=AH len=11 format=P content=
	//    ln=FULL-ADDRESS sn=A1 len=0 format=  content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=CITY sn=AJ len=20 format=A content=
	//    ln=ZIP sn=AK len=10 format=A content=
	//    ln=COUNTRY sn=AL len=3 format=A content=
	//    ln=TELEPHONE sn=A2 len=0 format=  content=
	//    ln=AREA-CODE sn=AN len=6 format=A content=
	//    ln=PHONE sn=AM len=15 format=A content=
	//    ln=DEPT sn=AO len=6 format=A content=
	//    ln=JOB-TITLE sn=AP len=25 format=A content=
	//    ln=INCOME sn=AQ len=0 format=  content=
	//    ln=CURR-CODE sn=AR len=3 format=A content=
	//    ln=SALARY sn=AS len=9 format=A content=
	//    ln=BONUS sn=AT len=9 format=A content=
	//    ln=BONUS sn=AT len=9 format=A content=
	//    ln=LEAVE-DATA sn=A3 len=0 format=  content=
	//    ln=LEAVE-DUE sn=AU len=2 format=N content=
	//    ln=LEAVE-TAKEN sn=AV len=2 format=N content=
	//    ln=LEAVE-BOOKED sn=AW len=0 format=  content=
	//    ln=LEAVE-START sn=AX len=8 format=A content=
	//    ln=LEAVE-END sn=AY len=8 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	// MAP Empl
	//   23 11
	//    ln=PERSONNEL-ID sn=AA len=8 format=A content=
	//    ln=FULL-NAME sn=AB len=0 format=  content=
	//    ln=FIRST-NAME sn=AC len=20 format=A content=
	//    ln=NAME sn=AE len=20 format=A content=
	//    ln=MIDDLE-I sn=AD len=1 format=A content=
	//    ln=MAR-STAT sn=AF len=1 format=A content=
	//    ln=SEX sn=AG len=1 format=A content=
	//    ln=BIRTH sn=AH len=4 format=D content=
	//    ln=FULL-ADDRESS sn=A1 len=0 format=  content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=CITY sn=AJ len=20 format=A content=
	//    ln=ZIP sn=AK len=10 format=A content=
	//    ln=COUNTRY sn=AL len=3 format=A content=
	//    ln=TELEPHONE sn=A2 len=0 format=  content=
	//    ln=AREA-CODE sn=AN len=6 format=A content=
	//    ln=PHONE sn=AM len=15 format=A content=
	//    ln=DEPT sn=AO len=6 format=A content=
	//    ln=JOB-TITLE sn=AP len=25 format=A content=
	//    ln=INCOME sn=AQ len=0 format=  content=
	//    ln=CURR-CODE sn=AR len=3 format=A content=
	//    ln=SALARY sn=AS len=9 format=A content=
	//    ln=BONUS sn=AT len=9 format=A content=
	//    ln=BONUS sn=AT len=9 format=A content=
	//    ln=LEAVE-DATA sn=A3 len=0 format=  content=
	//    ln=LEAVE-DUE sn=AU len=2 format=N content=
	//    ln=LEAVE-TAKEN sn=AV len=2 format=N content=
	//    ln=LEAVE-BOOKED sn=AW len=0 format=  content=
	//    ln=LEAVE-START sn=AX len=8 format=A content=
	//    ln=LEAVE-END sn=AY len=8 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=PHONETIC-NAME sn=PH len=20 format=A content=
	// MAP DublicateNames
	//   23 11
	//    ln=PERSONNEL-ID sn=AA len=8 format=A content=
	//    ln=FULL-NAME sn=AB len=21 format=  content=
	//    ln=FIRST-NAME sn=AC len=20 format=A content=
	//    ln=NAME sn=AE len=20 format=A content=
	//    ln=MIDDLE-I sn=AD len=1 format=A content=
	//    ln=MAR-STAT sn=AF len=1 format=A content=
	//    ln=SEX sn=AG len=1 format=A content=
	//    ln=BIRTH sn=AH len=4 format=D content=
	//    ln=FULL-ADDRESS sn=A1 len=50 format=  content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=ADDRESS-LINE sn=AI len=20 format=A content=
	//    ln=CITY sn=AJ len=20 format=A content=
	//    ln=ZIP sn=AK len=10 format=A content=
	//    ln=COUNTRY sn=AL len=3 format=A content=
	//    ln=TELEPHONE sn=A2 len=6 format=  content=
	//    ln=AREA-CODE sn=AN len=6 format=A content=
	//    ln=PHONE sn=AM len=15 format=A content=
	//    ln=DEPT sn=AO len=6 format=A content=
	//    ln=JOB-TITLE sn=AP len=25 format=A content=
	//    ln=INCOME sn=AQ len=12 format=  content=
	//    ln=CURR-CODE sn=AR len=3 format=A content=
	//    ln=SALARY sn=AS len=9 format=P content=
	//    ln=BONUS sn=AT len=9 format=P content=
	//    ln=BONUS sn=AT len=9 format=P content=
	//    ln=LEAVE-DATA sn=A3 len=2 format=  content=
	//    ln=LEAVE-DUE sn=AU len=2 format=N content=
	//    ln=LEAVE-TAKEN sn=AV len=2 format=N content=
	//    ln=LEAVE-BOOKED sn=AW len=8 format=  content=
	//    ln=LEAVE-START sn=AX len=8 format=N content=
	//    ln=LEAVE-END sn=AY len=8 format=N content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=LANG sn=AZ len=3 format=A content=
	//    ln=PH sn=PH len=0 format=A content=
	//    ln=LEAVE-LEFT sn=H1 len=0 format=A content=
	//    ln=DEPARTMENT sn=S1 len=0 format=A content=
	// MAP NewEmployees
	//   23 9
	//    ln=personnel-data sn=A0 len=0 format=  content=
	//    ln=personnel-id sn=AA len=8 format=A content=
	//    ln=id-data sn=AB len=0 format=  content=
	//    ln=personnel-no  !UQ taken! sn=AC len=4 format=I content=
	//    ln=id-card sn=AD len=8 format=B content=
	//    ln=signature sn=AE len=0 format=A content=
	//    ln=full-name sn=B0 len=0 format=  content=
	//    ln=first-name sn=BA len=40 format=U content=charset=UTF-8
	//    ln=middle-name sn=BB len=40 format=U content=charset=UTF-8
	//    ln=name sn=BC len=50 format=U content=charset=UTF-8
	//    ln=mar-stat sn=CA len=1 format=A content=
	//    ln=sex sn=DA len=1 format=A content=
	//    ln=birth sn=EA len=4 format=P content=
	//    ln=private-address sn=F0 len=0 format=  content=
	//    ln=address-line sn=FA len=60 format=A content=
	//    ln=address-line sn=FA len=60 format=A content=
	//    ln=city sn=FB len=40 format=A content=
	//    ln=post-code sn=FC len=10 format=A content=
	//    ln=country sn=FD len=3 format=A content=
	//    ln=phone-email sn=F1 len=0 format=  content=
	//    ln=area-code sn=FE len=6 format=A content=
	//    ln=private-phone sn=FF len=15 format=A content=
	//    ln=private-fax sn=FG len=15 format=A content=
	//    ln=private-mobile sn=FH len=15 format=A content=
	//    ln=private-email sn=FI len=80 format=A content=
	//    ln=private-email sn=FI len=80 format=A content=
	// MAP MF_TYPES_FRACTIONAL
	//   23 101
	//    ln=ISN sn=AA len=8 format=P content=
	//    ln=NUM_UNPACKED_W_SIGN_U sn=AB len=10 format=N content=
	//    ln=BINARY sn=AC len=10 format=B content=
	//    ln=B3_FIELD sn=AD len=3 format=B content=
	//    ln=ALPHA sn=AE len=32 format=A content=
	//    ln=PACKED_NUMERIC sn=AF len=10 format=P content=
	//    ln=B7_FIELD sn=AG len=7 format=B content=
	//    ln=B8_FIELD sn=AH len=8 format=B content=
	//    ln=B2_FIELD sn=AI len=2 format=B content=
	//    ln=P16_2 sn=AJ len=16 format=P content=
	//    ln=LOGICAL sn=AK len=1 format=L content=
	//    ln=NUM_UNPACKED_W_SIGN_N sn=AL len=10 format=N content=
	//    ln=FLOATING_POINT sn=AM len=8 format=F content=
	//    ln=B5_FIELD sn=AN len=5 format=B content=
	//    ln=ALPHA_NV_OPTION sn=AO len=32 format=A content=
	//    ln=INTEGER sn=AP len=4 format=I content=
	//    ln=B4_FIELD sn=AQ len=4 format=B content=
	//    ln=TIME sn=AR len=7 format=T content=
	//    ln=DATE sn=AS len=4 format=D content=
	//    ln=NUMERIC_UNPACKED_N sn=AT len=10 format=N content=
	//    ln=VARCHAR sn=AU len=0 format=A content=
	//    ln=B6_FIELD sn=AV len=6 format=B content=
	//    ln=PACKED_NUM_W_SIGN sn=AW len=10 format=P content=
	//    ln=B9_FIELD sn=AX len=9 format=B content=
	//    ln=B1_FIELD sn=AY len=1 format=B content=
	//    ln=NUMERIC_UNPACKED_U sn=AZ len=10 format=N content=
	// MAP MF_TYPES_FRACTIONAL2
	//   23 101
	//    ln=ISN sn=AA len=5 format=P content=
	//    ln=NUM_UNPACKED_W_SIGN_U sn=AB len=10 format=N content=
	//    ln=BINARY sn=AC len=10 format=B content=
	//    ln=B3_FIELD sn=AD len=3 format=N content=
	//    ln=ALPHA sn=AE len=32 format=A content=
	//    ln=PACKED_NUMERIC sn=AF len=6 format=N content=
	//    ln=B7_FIELD sn=AG len=7 format=N content=
	//    ln=B8_FIELD sn=AH len=8 format=B content=
	//    ln=B2_FIELD sn=AI len=2 format=N content=
	//    ln=P16_2 sn=AJ len=9 format=P content=fractionalshift=2
	//    ln=LOGICAL sn=AK len=1 format=N content=
	//    ln=NUM_UNPACKED_W_SIGN_N sn=AL len=10 format=N content=
	//    ln=FLOATING_POINT sn=AM len=8 format=F content=
	//    ln=B5_FIELD sn=AN len=5 format=N content=
	//    ln=ALPHA_NV_OPTION sn=AO len=32 format=A content=
	//    ln=INTEGER sn=AP len=4 format=N content=
	//    ln=B4_FIELD sn=AQ len=4 format=N content=
	//    ln=TIME sn=AR len=7 format=P content=
	//    ln=DATE sn=AS len=4 format=P content=
	//    ln=NUMERIC_UNPACKED_N sn=AT len=10 format=N content=
	//    ln=VARCHAR sn=AU len=0 format=A content=
	//    ln=B6_FIELD sn=AV len=6 format=N content=
	//    ln=PACKED_NUM_W_SIGN sn=AW len=6 format=N content=
	//    ln=B9_FIELD sn=AX len=9 format=B content=
	//    ln=B1_FIELD sn=AY len=1 format=N content=
	//    ln=NUMERIC_UNPACKED_U sn=AZ len=10 format=N content=
	// MAP Employees
	//   23 16
	//    ln=ID sn=AA len=8 format=A content=
	//    ln=FullName sn=AB len=40 format=  content=
	//    ln=FirstName sn=AC len=20 format=A content=
	//    ln=Name sn=AE len=20 format=A content=
	//    ln=MiddleName sn=AD len=20 format=A content=
	//    ln=MarriageState sn=AF len=1 format=A content=
	//    ln=Sex sn=AG len=1 format=A content=
	//    ln=Birth sn=AH len=4 format=D content=
	//    ln=Address sn=A1 len=50 format=  content=
	//    ln=AddressLine sn=AI len=20 format=A content=
	//    ln=AddressLine sn=AI len=20 format=A content=
	//    ln=City sn=AJ len=20 format=A content=
	//    ln=Zip sn=AK len=10 format=A content=
	//    ln=Country sn=AL len=3 format=A content=

}

func TestImportMaps(t *testing.T) {
	initTestLogWithFile(t, "mapjson.log")

	clearFile(4)
	maps, err := LoadJSONMap("Maps.json")
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Number of maps", len(maps))
	if assert.True(t, len(maps) > 0) {
		nrMaps := len(maps)
		err = maps[0].Store()
		assert.Error(t, err)

		for _, m := range maps {
			m.Repository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4}
			fmt.Println("MAP", m.Name)
			err = m.Store()
			if !assert.NoError(t, err) {
				return
			}
		}

		repo := NewMapRepositoryWithURL(DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4})
		ada, err := NewAdabas(NewURLWithDbid(adabasModDBID))
		if !assert.NoError(t, err) {
			return
		}
		maps, err = repo.LoadAllMaps(ada)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, maps, nrMaps)
		for _, m := range maps {
			err = m.Delete()
			if !assert.NoError(t, err) {
				return
			}
		}
		// Reinitiate repository
		repo = NewMapRepositoryWithURL(DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4})
		maps, err = repo.LoadAllMaps(ada)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, maps, 0)
		maps, err = LoadJSONMap("Maps.json")
		if !assert.NoError(t, err) {
			return
		}
		fmt.Println("Rewrite Number of maps", len(maps))
		for _, m := range maps {
			m.Repository = &DatabaseURL{URL: *NewURLWithDbid(adabasModDBID), Fnr: 4}
			fmt.Println("MAP", m.Name)
			err = m.Store()
			if !assert.NoError(t, err) {
				return
			}
		}
	}

}
