package main

import (
	//"errors"
	"fmt"
	//"bytes"
	//"strconv"
	//"strings"
	"encoding/json"
	//"regexp"
	//"bytes"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("GiouChaincode")


//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}


//==============================================================================================================================
//	Voter - Defines the structure for a voter object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type voter struct {
	//identity
	//ecert
	vID 						string `json:"vID"`
	tokensBought    			string `json:"tokensBought"`
	tokensUsedPerCandidate    string `json:"tokensUsedPerCandidate"` //Slice
}
/*
type Candidate struct {
	//identity
	//ecert
	cID 				string `json:"cID"`
	candidateName    string `json:"candidateName"`
	votesReceived    string `json:"votesReceived"`
}
*/

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}


// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}


//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================
/*
type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert string `json:"ecert"`
}
*/


//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function.
//==============================================================================================================================

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters() //ChaincodeStub func to call/invoke functions with parameters
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initVoter" { //create a new voter
		return t.initVoter(stub, args)
	} else if function == "deleteVoter" { //delete a voter
		return t.deleteVoter(stub, args)
	} else if function == "readVoter" { //read a voter
		return t.readVoter(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}


// ============================================================
// initVoter - create a new voter, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initVoter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2
	// "1", "1000", "0"
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init voter")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	/*if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}*/
	vID := args[0]
	tokensBought := args[1]
	tokensUsedPerCandidate := args[2]
	/*size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}
	*/
	// ==== Check if vID already exists ====
	voterAsBytes, err := stub.GetState(vID) //vID.voter?? & input=key=string gia Getstate
	if err != nil {
		return shim.Error("Failed to get vID: " + err.Error())
	} else if voterAsBytes != nil {
		fmt.Println("This vID already exists: " + vID)
		return shim.Error("This vID already exists: " + vID)
	}

	// ==== Create marble object and marshal to JSON ====
	//objectType := "marble"
	voter := &voter{vID, tokensBought, tokensUsedPerCandidate}
	voterJSONasBytes, err := json.Marshal(voter)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(vID, voterJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "vID~tokensBought"
	vIDtokensBoughtIndexKey, err := stub.CreateCompositeKey(indexName, []string{voter.vID, voter.tokensBought})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the voter.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(vIDtokensBoughtIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init voter")
	return shim.Success(nil)
}


// ===============================================
// readVoter - read a voter from chaincode state
// ===============================================
func (t *SimpleChaincode) readVoter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var vID, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the voter to query")
	}

	vID = args[0]
	valAsbytes, err := stub.GetState(vID) //get the voter from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + vID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Voter does not exist: " + vID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}


// ==================================================
// delete - remove a voter key/value pair from state
// ==================================================
func (t *SimpleChaincode) deleteVoter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var voterJSON voter
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	_vID := args[0]

	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetState(_vID) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + _vID + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + _vID + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &voterJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + _vID + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(_vID) //remove the marble from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "vID~tokensBought"
	vIDtokensBoughtIndexKey, err := stub.CreateCompositeKey(indexName, []string{voterJSON.vID, voterJSON.tokensBought})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(vIDtokensBoughtIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}


