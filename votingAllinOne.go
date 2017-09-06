/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"bytes"
	"strconv"
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


// ============================================================================================================================
// Asset Definitions - The ledger will store voters and candidates
// ============================================================================================================================

//==============================================================================================================================
//	Voter - Defines the structure for a voter object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Voter struct {
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


// ============================================================================================================================
// Init - initialize the chaincode 
//
// Marbles does not require initialization, so let's run a simple test instead.
//
// Shows off PutState() and how to pass an input argument to chaincode.
//
// Inputs - Array of strings
//  ["314"]
// 
// Returns - shim.Success or error
// ============================================================================================================================
// Init initializes chaincode
// ===========================

/*func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}*/

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("VotingApp Is Starting Up")
	_, args := stub.GetFunctionAndParameters()
	var Aval int
	var err error
	
	fmt.Println("Init() args count:", len(args))
	fmt.Println("Init() args found:", args)

	// expecting 1 arg for instantiate or upgrade
	if len(args) == 1 {
		fmt.Println("Init() arg[0] length", len(args[0]))

		// expecting arg[0] to be length 0 for upgrade
		if len(args[0]) == 0 {
			fmt.Println("args[0] is empty... must be upgrading")
		} else {
			fmt.Println("args[0] is not empty, must be instantiating")

			// convert numeric string to integer
			Aval, err = strconv.Atoi(args[0])
			if err != nil {
				return shim.Error("Expecting a numeric string argument to Init() for instantiate")
			}

			// this is a very simple test. let's write to the ledger and error out on any errors
			// it's handy to read this right away to verify network is healthy if it wrote the correct value
			err = stub.PutState("selftest", []byte(strconv.Itoa(Aval)))
			if err != nil {
				return shim.Error(err.Error())                  //self-test fail
			}
		}
	}

	// store compaitible Voting application version
	err = stub.PutState("voting_ui", []byte("4.0.0"))
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println(" - ready for action")                          //self-test pass
	return shim.Success(nil)
}


//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function.
//==============================================================================================================================
// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "init" {                    //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" {             //generic read ledger -> psiloaxrhsth
		return read(stub, args)
	} else if function == "write" {            //generic writes to ledger -> psiloaxrhsth
		return write(stub, args)
	} else if function == "delete_marble" {    //deletes a marble from state
		return delete_marble(stub, args)
	} else if function == "init_voter" {      //create a new marble
		return init_voter(stub, args)
	} else if function == "set_owner" {        //change owner of a marble
		return set_owner(stub, args)
	} else if function == "init_owner"{        //create a new marble owner
		return init_owner(stub, args)
	} else if function == "read_everything"{   //read everything, (owners + marbles + companies)
		return read_everything(stub)
	} else if function == "getHistory"{        //read history of a marble (audit)
		return getHistory(stub, args)
	} else if function == "getMarblesByRange"{ //read a bunch of marbles by start and stop id
		return getMarblesByRange(stub, args)
	} else if function == "disable_owner"{     //disable a marble owner from appearing on the UI
		return disable_owner(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}


//********************************** WRITE LEDGER **********************************
// ============================================================================================================================
// Init Owner - create a new owner aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      owner id   , username, company
// "o9999999999999",     bob", "united marbles"
// ============================================================================================================================
func init_voter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_owner")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var voter Voter
	//voter.ObjectType = "marble_owner"
	voter.vID =  args[0]
	voter.tokensBought = args[1]
	voter.tokensUsedPerCandidate = args[2]
	//owner.Enabled = true
	fmt.Println(voter)

	//check if user already exists
	_, err = get_voter(stub, voter.vID)
	//h get_voter an uparxei hdh o voter epistrefei nill, dld uparxei error
	if err == nil {
		fmt.Println("This owner already exists - " + voter.vID)
		return shim.Error("This owner already exists - " + voter.vID)
	}

	//store user
	voterAsBytes, _ := json.Marshal(voter)                         //convert to array of bytes
	err = stub.PutState(voter.vID, voterAsBytes)                    //store owner by its Id
	if err != nil {
		fmt.Println("Could not store user")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_voter marble")
	return shim.Success(nil)
}


// ============================================================================================================================
// delete_marble() - remove a marble from state and from marble index
// 
// Shows Off DelState() - "removing"" a key/value from the ledger
//
// Inputs - Array of strings
//      0      ,         1
//     id      ,  authed_by_company
// "m999999999", "united marbles"
// ============================================================================================================================
func delete_voter(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	//might be need to provide the var voter Voter 
	fmt.Println("starting delete_marble")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	vid := args[0]
	//authed_by_company := args[1]

	// get the marble
	voter, err := get_voter(stub, vid)
	if err != nil{
		fmt.Println("Failed to find marble by vid " + vid)
		return shim.Error(err.Error())
	}

	// check authorizing company (see note in set_owner() about how this is quirky)
	//if marble.Owner.Company != authed_by_company{
	//	return shim.Error("The company '" + authed_by_company + "' cannot authorize deletion for '" + marble.Owner.Company + "'.")
	//}

	if voter.vID != vid {                                     //test if marble is actually here or just nil
		return shim.Error("Not the same voter ID provided")	//the existance of the voter is checked in the get_voter func
	}

	// remove the marble
	err = stub.DelState(vid)                                                 //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_marble")
	return shim.Success(nil)
}


//********************************** READ LEDGER **********************************
// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
// 
// Returns - string
// ============================================================================================================================
func read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var voter Voter
	var jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	/*
	voter.vID = args[0]
	vid = args[0]
	valAsbytes, err := stub.GetState(vid)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}
	*/
	
	// get the marble
	voter, err := get_voter(stub, vid)
	if err != nil{
		fmt.Println("Failed to find marble by vid " + vid)
		return shim.Error(err.Error())
	}

	fmt.Println("- end read")
	//return shim.Success(valAsbytes)
	return shim.Success(voter)                  //send it onward
}


//********************************** LIB **********************************
// ============================================================================================================================
// Get Marble - get a marble asset from ledger
//vid is the input
//vID is from struct Voter
// ============================================================================================================================
func get_voter(stub shim.ChaincodeStubInterface, vid string) (Voter, error) {
	var voter Voter
	voterAsBytes, err := stub.GetState(vid)                  //getState retreives a key/value from the ledger

	if err != nil {                                          //this seems to always succeed, even if key didn't exist <<<<----------------------------------------------------------------------
		return voter, errors.New("Failed to find voter - " + vid)
	}
	json.Unmarshal(voterAsBytes, &voter)                   //un stringify it aka JSON.parse()

	if voter.vID != vid {                                     //test if marble is actually here or just nil
		return voter, errors.New("Voter does not exist - " + vid)
	}

	return voter, nil
}


// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
//didn't change it, seems ok
// ========================================================
func sanitize_arguments(strs []string) error{
	for i, val := range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		if len(val) > 32 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		}
	}
	return nil
}
