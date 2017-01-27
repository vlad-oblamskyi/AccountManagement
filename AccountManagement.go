package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	b64 "encoding/base64"
	"time"
	"encoding/json"
	"github.com/hyperledger/fabric/core/util"
)

const KVS_HANLDER_KEY = "KVS_HANDLER_KEY"

type AccountManagement struct {
}

type AccountKey struct {
	HolderBIC string  `json:"holderBic"`
	OwnerBIC  string  `json:"ownerBic"`
	Currency  string  `json:"currency"`
	Type      string  `json:"type"`
}

type AccountValue struct {
	Amount    string  `json:"amount"`
	Currency  string  `json:"currency"`
	Type      string  `json:"type"`
	Date      string  `json:"date"`
	Number    string  `json:"number"`
}

type UserKey struct {
	BIC    string  `json:"bic"`
	Login  string  `json:"login"`
}

type PermissionAccountKey struct {
	Type         string  `json:"type"`
	Holder       string  `json:"holder"`
	Owner        string  `json:"owner"`
	Currency     string  `json:"currency"`
	AccountType  string  `json:"accountType"`
}

type Permission struct {
	Key      PermissionAccountKey  `json:"accountKey"`
	Access   string  `json:"access"`
}

type UserDetails struct {
	Password     string  `json:"password"`
	Permissions  []Permission  `json:"permissions"`
}

func (t *AccountManagement) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. KVS chaincode id is expected");
	}
	stub.PutState(KVS_HANLDER_KEY, []byte(args[0]))

	return nil, nil
}

func (t *AccountManagement) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "addAccount":
		if len(args) != 6 {
			return nil, errors.New("Incorrect number of arguments. Expecting 6: holder bic, owner bic, number, amount, currency, type");
		}

		dateTime := time.Now().UTC()

		accountKey := &AccountKey {
			HolderBIC: args[0],
			OwnerBIC: args[1],
			Currency: args[4],
			Type: args[5],
		}
		accountValue := &AccountValue {
			Amount: args[3],
			Currency: args[4],
			Type: args[5],
			Date: dateTime.Format(time.RFC3339),
			Number: args[2],
		}
		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		jsonAccountKey, _ := json.Marshal(accountKey)
		jsonAccountValue, _ := json.Marshal(accountValue)
		invokeArgs := util.ToChaincodeArgs("put", jsonAccountKey, jsonAccountValue)

		stub.InvokeChaincode(mapId, invokeArgs)
		return nil, nil
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func (t *AccountManagement) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "listAccounts":
		if len(args) != 1 {
			return nil, errors.New("Incorrect number of arguments. Expecting auth token of the authenticated user to query");
		}

		jsonUserKey, _ := b64.StdEncoding.DecodeString(string(args[0]))
		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		queryArgs := util.ToChaincodeArgs("function", string(jsonUserKey))

		queryResult, _ := stub.QueryChaincode(mapId, queryArgs)
		var userDetails UserDetails
		if err := json.Unmarshal(queryResult, &userDetails); err != nil {
			panic(err)
		}

		accounts := make([]string, 0)
		for i := 0; i < len(userDetails.Permissions); i ++ {
			accountKey := &AccountKey{
				HolderBIC: userDetails.Permissions[i].Key.Holder,
				OwnerBIC: userDetails.Permissions[i].Key.Owner,
				Currency: userDetails.Permissions[i].Key.Currency,
				Type: userDetails.Permissions[i].Key.Type,
			}
			invokeArgs := util.ToChaincodeArgs("function", json.Marshal(accountKey))
			account, _ := stub.InvokeChaincode(mapId, invokeArgs)
			if account != nil {
				accounts = append(accounts, account)
			}
		}
		return json.Marshal(accounts)
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func main() {
	err := shim.Start(new(AccountManagement))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}