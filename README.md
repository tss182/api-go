# Install
    go get github.com/tss182/api

# How to using

    apiCfg := api.Init{}
	//Set URL
	apiCfg.Url = "https://api.github.com/users/tss182/repos"
    
    // set Conten type
    apiCfg.ContentType = api.TypeMultipart
    
    //set Method
    apiCfg.Method = api.MethodPOST
    
	//Set Body
	/*apiCfg.Body = map[string]interface{}{
	    "id": 12321,
	    "name": "John",
	}*/
	// Content Type api.TypeJson in body can input struct with json flag

	//Set Header
	/*apiCfg.Data = map[string]interface{}{
        "api-key": "1231231231111",
    }*/

	//Proccess
	err := apiCfg.Do()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//get respon with struct
	var result []struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	}
	err = apiCfg.Get(&result)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//get raw respon
	fmt.Println(apiCfg.GetRaw())

	//print result
	fmt.Println(result)



Respon support JSON and XML
