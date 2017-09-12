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
	//"bytes"
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
	//tokensUsedPerCandidate    string `json:"tokensUsedPerCandidate"` //Map oxi Slice!
	//tokensUsedPerCandidate 		make(map[string]string) `json:"tokensUsedPerCandidate"`
	tokensUsedPerCandidate 		map[string]string `json:"tokensUsedPerCandidate"`
	tokensRemaining				string `json:"tokensRemaining"`
	Enabled						bool `json:"Enabled"`
}

type Candidate struct {
	//identity
	//ecert
	cID 				string `json:"cID"`
	candidateName    string `json:"candidateName"`
	votesReceived    string `json:"votesReceived"`
}


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
	fmt.Println("\nVotingApp Is Starting Up\n")
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

	fmt.Println("\n - ready for action")                          //self-test pass
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
	} else if function == "read_voter" {             //generic read ledger -> psiloaxrhsth
		return read_voter(stub, args)
	} else if function == "delete_voter" {    //deletes a marble from state
		return delete_voter(stub, args)
	} else if function == "init_voter" {      //create a new marble
		return init_voter(stub, args)
	}else if function == "init_candidate" {      //create a new marble
		return init_candidate(stub, args)
	}else if function == "read_candidate" {      //create a new marble
		return read_candidate(stub, args)
	}else if function == "delete_candidate" {      //create a new marble
		return delete_candidate(stub, args)
	}else if function == "disable_voter" {      //create a new marble
		return disable_voter(stub, args)
	}else if function == "transfer_vote" {      //create a new marble
		return transfer_vote(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

//*********************************************************************************
//********************************** WRITE LEDGER *********************************
//*********************************************************************************
// ============================================================================================================================
// Init Owner - create a new owner aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      owner id   , username, company
// "o9999999999999",     bob", "united marbles"
//
// Inputs - Array of Strings
//           0     ,         1   ,   2
//      voter id   , tokensBought, company
//           "v001",       "100" , "united marbles"
// ============================================================================================================================
func init_voter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_voter")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var voter Voter
/*	//voter.ObjectType = "marble_owner"
	voter.vID =  args[0]
	voter.tokensBought = args[1]
	//voter.tokensUsedPerCandidate = args[2]
	voter.tokensRemaining = args[1]
	voter.Enabled = true
	fmt.Println(voter)
*/
	_vid :=  args[0]
	_tokensBought := args[1]
	//voter.tokensUsedPerCandidate = args[2]
	_tokensRemaining := args[1]
	_Enabled := true
	fmt.Println("ID: " + _vid + " tokensBought: " + _tokensBought + " tokensRemaining: " + _tokensRemaining + " Active: " + strconv.FormatBool(_Enabled))

	//check if user already exists
	voter, err = get_voter(stub, _vid)
	//h get_voter an uparxei hdh o voter epistrefei nill, dld uparxei error <- I've changed that
	if err == nil {
		fmt.Println("This voter already exists - " + _vid)
		return shim.Error("This voter already exists - " + _vid)
	}

	//store user
	voterAsBytes, _ := json.Marshal(voter)                         //convert to array of bytes
	fmt.Println(" putting state in block")
	err = stub.PutState(voter.vID, voterAsBytes)                    //store owner by its Id
	fmt.Println(voter.vID + " voter has been stored")
	if err != nil {
		fmt.Println("Could not store voter")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_voter")
	return shim.Success(nil)
}


func init_candidate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_candidate")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var candidate Candidate
	//voter.ObjectType = "marble_owner"
	candidate.cID =  args[0]
	candidate.candidateName = args[1]
	//candidate.votesReceived = args[2]
	fmt.Println(candidate)

	//check if user already exists
	candidate, err = get_candidate(stub, candidate.cID)
	//h get_voter an uparxei hdh o voter epistrefei nill, dld uparxei error
	//apo == to ekane != gt dn leitourgouse swsta.check it again
	if err != nil {
		fmt.Println("This candidate already exists - " + candidate.cID)
		return shim.Error("This candidate already exists - " + candidate.cID)
	}

	//store user
	candidateAsBytes, _ := json.Marshal(candidate)                         //convert to array of bytes
	err = stub.PutState(candidate.cID, candidateAsBytes)                    //store owner by its Id
	fmt.Println(candidate.cID + " candidate has been stored")
	if err != nil {
		fmt.Println("Could not store candidate")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_candidate")
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
	fmt.Println("starting delete_voter")

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
		fmt.Println("Failed to find voter by vid " + vid)//leitourgei, alla to evgale kai gia voter pou uphrxe.
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

	fmt.Println("- end delete_voter")
	return shim.Success(nil)
}


func delete_candidate(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	//might be need to provide the var voter Voter  -> NOT
	fmt.Println("starting delete_candidate")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	cid := args[0]
	//authed_by_company := args[1]

	// get the marble
	candidate, err := get_candidate(stub, cid)
	if err != nil{
		fmt.Println("Failed to find candidate by cid " + cid)
		return shim.Error(err.Error())
	}

	// check authorizing company (see note in set_owner() about how this is quirky)
	//if marble.Owner.Company != authed_by_company{
	//	return shim.Error("The company '" + authed_by_company + "' cannot authorize deletion for '" + marble.Owner.Company + "'.")
	//}

	if candidate.cID != cid {                                     //test if marble is actually here or just nil
		return shim.Error("Not the same candidate ID provided")	//the existance of the voter is checked in the get_voter func
	}

	// remove the marble
	err = stub.DelState(cid)                                                 //remove the key from chaincode state
	if err != nil {
		fmt.Println("Failed to delete candidate by cid " + cid)
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_candidate")
	return shim.Success(nil)
}


// ============================================================================================================================
// Disable Marble Owner
//
// Shows off PutState()
//
// Inputs - Array of Strings
//       0     ,        1      
//  owner id       , company that auth the transfer
// "o9999999999999", "united_mables"
// ============================================================================================================================
func disable_voter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting disable_voter")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var vid = args[0]
	//var authed_by_company = args[1]

	// get the marble owner data
	voter, err := get_voter(stub, vid)
	// this if might duplicating if in the get_voter func
	if err != nil {
		fmt.Println("This voter does not exist - " + vid)
		return shim.Error("This voter does not exist - " + vid)//monimos auto printarei
	}

	// check authorizing company
	//if owner.Company != authed_by_company {
	//	return shim.Error("The company '" + authed_by_company + "' cannot change another companies marble owner")
	//}

	// disable the owner
	//duplicate if in transfer_vote
	tR, err := strconv.Atoi(voter.tokensRemaining)
	if tR <= 0 {
		fmt.Println(" Voter - " + vid + " - is gonna be disabled because of not remaining tokens")
		voter.Enabled = false
		jsonAsBytes, _ := json.Marshal(voter)         //convert to array of bytes
		err = stub.PutState(args[0], jsonAsBytes)     //rewrite the owner
		if err != nil {
			return shim.Error(err.Error())
		}

		fmt.Println("- end disable_voter")
		return shim.Success(nil)
	}

	return shim.Error("The voter '" + vid + "' has " + voter.tokensRemaining + " remaining tokens")
}


// ============================================================================================================================
// Set Owner on Marble
//
// Shows off GetState() and PutState()
//
// Inputs - Array of Strings
//       0     ,        1      ,        2
//  marble id  ,  to owner id  , company that auth the transfer
// "m999999999", "o99999999999", united_mables" 
// ============================================================================================================================
func transfer_vote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var voter Voter
	var candidate Candidate
	var err error
	fmt.Println("starting transfer_vote")

	// this is quirky
	// todo - get the "company that authed the transfer" from the certificate instead of an argument
	// should be possible since we can now add attributes to the enrollment cert
	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	vid := args[0]
	//cName := args[1]
	cid := args[1]
	tokensToUse := args[2]
	fmt.Println("The voter " + vid + " votes for the candidate " + cid + " with the amount of- |" + tokensToUse + "| -tokens.")

	//check if user already exists
	voter, err = get_voter(stub, vid)//change to vid
	//h get_voter an uparxei hdh o voter epistrefei nill, dld uparxei error
	if err != nil || voter.Enabled == false {
		return shim.Error("This voter does not exist or is disabled- " + voter.vID)//change to vid
	}

	//check if user already exists
	candidate, err = get_candidate(stub, cid)
	if err != nil {
		return shim.Error("This candidate does not exist - " + cid)//cid
	}

	//var tR *int
	tR, err := strconv.Atoi(voter.tokensRemaining)
	tTU, err := strconv.Atoi(tokensToUse)
	vR, err := strconv.Atoi(candidate.votesReceived)

	if (tR >= tTU) {
		tR = tR - tTU
		vR = vR + tTU
		voter.tokensRemaining = strconv.Itoa(tR)
		fmt.Println("the voter's remaining tokens are " + voter.tokensRemaining)
        voter.tokensUsedPerCandidate[cid] = tokensToUse
	}else if (tR > 0 && tTU >tR) {
		fmt.Printf("Not enough tokens. Your maximum amount of tokens is: - |" + voter.tokensRemaining + "| -")
	}else if (tR <= 0) {
		var v = []string {vid}
		fmt.Printf("The voter with vid " + vid + " is gonna be disabled")
		_ = disable_voter(stub, v)
	}else{
		fmt.Printf("None of the values is matching\n" )
	}

	// get marble's current state
/*	marbleAsBytes, err := stub.GetState(marble_id)
	if err != nil {
		return shim.Error("Failed to get marble")
	}
	res := Marble{} //res = response
	json.Unmarshal(marbleAsBytes, &res)           //un stringify it aka JSON.parse()

	// check authorizing company
	if res.Owner.Company != authed_by_company{
		return shim.Error("The company '" + authed_by_company + "' cannot authorize transfers for '" + res.Owner.Company + "'.")
	}
	
	// transfer the marble
	res.Owner.Id = new_owner_id                   //change the owner
	res.Owner.Username = owner.Username
	res.Owner.Company = owner.Company
	jsonAsBytes, _ := json.Marshal(res)           //convert to array of bytes
	err = stub.PutState(args[0], jsonAsBytes)     //rewrite the marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}
*/
	//store user
	voterAsBytes, _ := json.Marshal(voter)
	err = stub.PutState(voter.vID, voterAsBytes)
	if err != nil{
		fmt.Println("Could not store voter")
		return shim.Error(err.Error())
	}

	//store user
	candidateAsBytes, _ := json.Marshal(candidate)                         //convert to array of bytes
	err = stub.PutState(candidate.cID, candidateAsBytes)                    //store owner by its Id
	if err != nil {
		fmt.Println("Could not store candidate")
		return shim.Error(err.Error())
	}

	fmt.Println("- end transfer_vote")
	return shim.Success(nil)
}


//*********************************************************************************
//********************************** READ LEDGER **********************************
//*********************************************************************************
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
func read_voter(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error
	fmt.Println("starting read_voter")

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
/*
	// getthe marble
	vid := args[0]
	_, err := get_voter(stub, vid)
	if err != nil{
		fmt.Println("Failed to find marble by vid " + vid)
		return shim.Error(err.Error())
	}
*/
	vid := args[0]
	voterAsbytes, err := stub.GetState(vid)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + vid + "\"}"
		return shim.Error(jsonResp)
	}
	fmt.Println("read was successful")
	var voter Voter
	json.Unmarshal(voterAsbytes, &voter)
	fmt.Println(voter)
	fmt.Println("- end read")
	//return shim.Success(valAsbytes)
	return shim.Success(voterAsbytes)                  //send it onward
}


func read_candidate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var candidate Candidate //prostethike gia to json.Unmarshal etsi wste na ginetai print swsto.
	var jsonResp string
	var err error
	fmt.Println("starting read candidate")

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
/*
	// getthe marble
	vid := args[0]
	_, err := get_voter(stub, vid)
	if err != nil{
		fmt.Println("Failed to find marble by vid " + vid)
		return shim.Error(err.Error())
	}
*/
	cid := args[0]
	candidateAsbytes, err := stub.GetState(cid)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + cid + "\"}"
		return shim.Error(jsonResp)
	}
	json.Unmarshal(candidateAsbytes, &candidate)
	fmt.Println("the candidate: -| " + candidate.cID + " |- ")//den leitourgei
	fmt.Println("read was successful")
	fmt.Println("- end read")
	//return shim.Success(valAsbytes)
	return shim.Success(candidateAsbytes)                  //send it onward
}

//*********************************************************************************
//********************************** LIB ******************************************
//*********************************************************************************
// ============================================================================================================================
// Get Marble - get a marble asset from ledger
//vid is the input
//vID is from struct Voter
//estw v001=voter.vID->init vid=v001->GetState(001)->unmarshal(001) &voter(001)->001!=001 ?
//voter.vID != vid -> Dld elegxei an h domh me 001 einai diaforetikh me 001??auto den einai anousio?
// ============================================================================================================================
func get_voter(stub shim.ChaincodeStubInterface, vid string) (Voter, error) {
	var voter Voter
	voterAsBytes, err := stub.GetState(vid)                  //getState retreives a key/value from the ledger

	//err == true
	if err != nil {                                          //this seems to always succeed, even if key didn't exist <<<<----------------------------------------------------------------------
		fmt.Println("Voter does not exist - " + vid)  //test if marble is actually here or just nil
		json.Unmarshal(voterAsBytes, &voter)
		return voter, errors.New("Voter does not exist - " + vid)
	} else {
		fmt.Println("Voter exists - " + vid)
		return voter, nil //
		//, errors.New("Voter exists - " + vid)
	}

/*	if err != nil {                                          //this seems to always succeed, even if key didn't exist <<<<----------------------------------------------------------------------
		return voter, errors.New("Failed to find voter - " + vid)
	}
	json.Unmarshal(voterAsBytes, &voter)                   //un stringify it aka JSON.parse()

	//auto to if fainetai axrhsto
	if voter.vID != vid {  
		fmt.Println("Voter does not exist - " + vid)  //test if marble is actually here or just nil
		return voter, nil // leitourgei otan kanw delete egguro/akuro vID
	}
*/
	
}


func get_candidate(stub shim.ChaincodeStubInterface, cid string) (Candidate, error) {
	var candidate Candidate
	candidateAsBytes, err := stub.GetState(cid)                  //getState retreives a key/value from the ledger

	if err != nil {                                          //this seems to always succeed, even if key didn't exist <<<<----------------------------------------------------------------------
		return candidate, errors.New("Failed to find candidate - " + cid)
	}
	json.Unmarshal(candidateAsBytes, &candidate)                   //un stringify it aka JSON.parse()

	if candidate.cID != cid {
		fmt.Println("Candidate does not exist - " + cid)      //test if marble is actually here or just nil
		return candidate, nil 
	}

	return candidate, errors.New("Candidate exists - " + cid)
}


// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
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
