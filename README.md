# Install
    go get github.com/tss182/api

# How to using

    package main
    
    import (
    	"fmt"
    	"github.com/tss182/api-go"
    )
    
    func main()  {
    	apiCfg := api.Api{}
    	//Set URL
    	apiCfg.Url = "https://api.github.com/users/tss182/repos"
    
    	// set Conten type
    	apiCfg.ContentType = api.TypeUrlEncode
    
    	//set Method
        apiCfg.Method = api.MethodGET
    
    	//Set Body
    	/*apiCfg.Body = map[string]interface{}{
    	    "id": 12321,
    	    "name": "John",
    	}*/
    	// Content Type api.TypeJson in body can input struct with json flag
    
    	//Set Header
        //apiCfg.HeaderAdd("api-key","23423234")
        //apiCfg.HeaderAdd("time",time.Now().String())
    
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
    	
    	//print header respon code
    	fmt.Println(api.Status)
    }




Respon support JSON and XML
