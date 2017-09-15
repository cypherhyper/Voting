----
# VOTING CHAINCODE DEPLOY
Execute the following instructions serial.

----
## Vagrant dev env
[Instructions](http://hyperledger-fabric.readthedocs.io/en/v1.0.0-beta/dev-setup/devenv.html)


----
First of all log in vagrant:
* `cd ~/go/src/github.com/hyperledger/fabric/devenv/`
* `vagrant up`
* `vagrant ssh`


----
Download fabric-sample
* Create a directory where you want to. `mkdir fabric-sample`
* `cd fabric-sample`
* Execute : `curl -sSL https://goo.gl/LQkuoh | bash`


----
## Deploy


* `cd release/linux-amd64$`
* `./network_setup.sh down` - needed in case it's not the first time deploying.
* `./generateArtifacts.sh mychannel`
* `vi docker_compose_cli.yaml` - needed in case it's the first time deploying.

	And comment the following line: `command: /bin/bash -c './scripts/script.sh ${CHANNEL_NAME}; sleep $TIMEOUT'`. Your file should look like this:
	````
	working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
	# command: /bin/bash -c './scripts/script.sh ${CHANNEL_NAME}; sleep $TIMEOUT'
	volumes
	````
* `CHANNEL_NAME=mychannel TIMEOUT=3600 docker-compose -f docker-compose-cli.yaml up -d`

* `docker exec -it cli bash`

* `peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem`

* `peer channel join -b mychannel.block`

* `go get github.com/giou-k/Voting`

* `peer chaincode install -n mycc -v 1.0 -p github.com/giou-k/Voting`

* `peer chaincode instantiate -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -v 1.0 -p github.com/giou-k/Voting -c '{"Args":[]}' -P "OR ('Org1MSP.member','Org2MSP.member')"`

* After the instantiation you can see logs. 

----
## Logs

* `docker ps`
* `docker logs -f <container id>` . The `<container id>` will have as `<image>` yours chaincode name (ex: mycc) **after** the instantiation of the CC.


----
## Test Init

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["init","314"]}'`


----
## Voter Invoke - Query -Delete

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["init_voter","v001","100"]}'`

* `peer chaincode query -C mychannel -n mycc -c '{"Args":["read_voter","v001"]}'`

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["delete_voter","v001"]}'`


----
## Candidate Invoke - Query - Delete

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["init_candidate","c001","christopher wallace"]}'`

* `peer chaincode query -C mychannel -n mycc -c '{"Args":["read_candidate","c001"]}'`

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["delete_candidate","c001"]}'`


----
## Transfer Vote

* `peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/cacerts/ca.example.com-cert.pem -C mychannel -n mycc -c '{"Args":["transfer_vote","v001","c001","20"]}'`



