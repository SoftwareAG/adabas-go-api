package adabas

func  ExampleReadRequest_fileMf() {
		initLogWithFile("mainframe.log")
		network := os.Getenv("ADAMFDBID")
		if network == "" {
			return
		}
			connection, cerr := NewConnection("acj;target="+network)
		if cerr != nil {
			return
		}
		defer connection.Close()
		request, err := connection.CreateReadRequest(1)
		if err != nil {
			fmt.Println("Error read map : ", err)
			return
		}
		fmt.Println("Connection : ", connection)
	
		fmt.Println("Limit query data:")
		request.QueryFields("AA,AB")
		request.Limit = 2
		result := &RequestResult{}
		fmt.Println("Read logical data:")
		err = request.ReadLogicalWithWithParser("AA=[11100301:11100303]", nil, result)
		if err != nil {
			fmt.Println("Error reading", err)
			return
		}
		fmt.Println("Result data:")
		result.DumpValues()
		// Output: Connection :  Adabas url=23 fnr=0
		// Limit query data:
		// Read logical data:
		// Result data:
		// Dump all result values
		// Record Isn: 0251
		//   AA = > 11100301 <
		//   AB = [ 1 ]
		//    AC = > HANS                 <
		//    AE = > BERGMANN             <
		//    AD = > WILHELM              <
		// Record Isn: 0383
		//   AA = > 11100302 <
		//   AB = [ 1 ]
		//    AC = > ROSWITHA             <
		//    AE = > HAIBACH              <
		//    AD = > ELLEN                <
	
}