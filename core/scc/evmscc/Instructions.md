Installing the EVMSCC with Fabric:

1. Use Fabric with `evmscc` from `https://github.com/swetharepakula/fabric/tree/master`. This is based off `fabric master`.
1. The  `sampleconfig/core.yml` being used to standup the network has already enabled `evmscc` (If using repo above).
```
    # system chaincodes whitelist. To add system chaincode "myscc" to the
    # whitelist, add "myscc: enable" to the list below, and register in
    # chaincode/importsysccs.go
    system:
        cscc: enable
        lscc: enable
        escc: enable
        vscc: enable
        qscc: enable
        evmscc: enable
```

Install & Interact SimpleStorage Smart Contract:
1. `SimpleStorage` is the following contract:
```
  pragma solidity ^0.4.0;

  contract SimpleStorage {
      uint storedData;

      function set(uint x) public {
          storedData = x;
      }

      function get() public constant returns (uint) {
          return storedData;
      }
  }
```
```
compiledBytecode :
 6060604052341561000f57600080fd5b60d38061001d6000396000f3006060604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806360fe47b114604e5780636d4ce63c14606e575b600080fd5b3415605857600080fd5b606c60048080359060200190919050506094565b005b3415607857600080fd5b607e609e565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820e170013eadb8debdf58398ee9834aa86cf08db2eee5c90947c1bcf6c18e3eeff0029
```
1. To deploy a contract your args should be:
```
peer chaincode invoke -C <channelname> -n evmscc -c {"Args":["0000000000000000000000000000000000000000","<bytecode>"]}
```
The payload of that is the contract address.
1. To interact with your contract your args will always follow this pattern:
```
{"Args":["<contract address>","<functionhash + arg>"]}
```
For SimpleStorage here are function hashes:
 - SET: `60fe47b1`  -> Takes one 32 byte argument Ex: To set the value as one, I would use: `60fe47b10000000000000000000000000000000000000000000000000000000000000001`
 - GET: `6d4ce63c` -> **NOTE** For GET, remember to add the --hex option so the output is readable.
1. Always use `peer chaincode invoke` for functions that write to the ledger. In the example the last 20 bytes of the second argument is for SET. Here we are setting the value to 1
```
peer chaincode invoke -C <channelname> -n evmscc -c {"Args":["<contract address","60fe47b10000000000000000000000000000000000000000000000000000000000000001"]}
```
1. Always use `peer chaincode query` for functions that do not write to the ledger.
```
peer chaincode query -C <channelname> -n evmscc -c {"Args":["<contract address", "6d4ce63c"]} --hex
```
