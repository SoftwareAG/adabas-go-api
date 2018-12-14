package adatypes

import (
	"fmt"
)
func ExampleDefinition_ShouldRestrictToFields() {
	f, err := initLogWithFile("definition.log")
	if err != nil {
		return
	}
	defer f.Close()
	initDefinitionCache()
	testDefinition := createPeriodGroupMultiplerField()
	testDefinition.PutCache("AA")
	testDefinition.DumpTypes(false, false)
	testDefinition.DumpTypes(false, true)
	err = testDefinition.ShouldRestrictToFields("GC,I8")
	if err != nil {
		fmt.Println("Restrict original entry", err)
	}
	definition := CreateDefinitionByCache("AA")
	if definition == nil {
		fmt.Println("Error create cache definition nil")
	}
	err = definition.ShouldRestrictToFields("GC,I8")
	if err != nil {
		fmt.Println("Restrict cached entry error", err)
	}
	definition.DumpTypes(false, false)
	definition.DumpTypes(false, true)

	// Output: Dump all file field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
	//
	// Dump all active field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
	//
	// Dump all file field types:
	//   1, U4, 4, B  ; U4  PE=false MU=false REMOVE=true
	//   1, B1, 1, F  ; B1  PE=false MU=false REMOVE=true
	//   1, UB, 1, B  ; UB  PE=false MU=false REMOVE=true
	//   1, I2, 2, B  ; I2  PE=false MU=false REMOVE=true
	//   1, U8, 8, B  ; U8  PE=false MU=false REMOVE=true
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=true
	//     2, GM, 5, P ,MU; GM  PE=true MU=true REMOVE=true
	//       3, GM, 5, P  ; GM  PE=true MU=true REMOVE=true
	//     2, GS, 1, A  ; GS  PE=true MU=true REMOVE=true
	//     2, GP, 1, P  ; GP  PE=true MU=true REMOVE=true
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=true
	//
	// Dump all active field types:
	//   1, GR ,PE ; GR  PE=true MU=true REMOVE=true
	//     2, GC, 1, A  ; GC  PE=true MU=true REMOVE=false
	//   1, I8, 8, B  ; I8  PE=false MU=false REMOVE=false
}
