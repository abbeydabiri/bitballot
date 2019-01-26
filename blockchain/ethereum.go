package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	// "github.com/ethereum/go-ethereum/crypto/sha3"

	"bitballot/config"
)

var err error
var client *ethclient.Client
const WEI = 1000000000000000000

var ETHAddress = 0


//EthClientDial ...
func EthClientDial() {
	client, err = ethclient.Dial(config.Get().Ethereum.Network)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("we have a connection to Infura network: " + config.Get().Ethereum.Network)
	_ = client
}

//ETHNewMnemonic ...
func ETHNewMnemonic() string {
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}

//ETHIsMnemonicValid ...
func ETHIsMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

//EthGenerateKey ...
func EthGenerateKey(level int) (privateKey *ecdsa.PrivateKey, fromAddress common.Address)  {
	
	seed := bip39.NewSeed(config.Get().Ethereum.Mnemonic, "")
	masterPrivateKey, _ := bip32.NewMasterKey(seed)
	masterPublicKey := masterPrivateKey.PublicKey()

	
	// Display mnemonic and keys
	fmt.Printf("Mnemonic: [%v] \n",config.Get().Ethereum.Mnemonic)
	fmt.Println("Master private key: ", masterPrivateKey)
	fmt.Println("Master public key: ", masterPublicKey.PublicKey())

	const Purpose uint32 = 0x8000002C
	const CoinEther uint32 = 0x8000003c
	const Account uint32 = 0x80000000
	const SubAccounts uint32 = 0x00000000 //this must be uint32 0

	child, err := masterPrivateKey.NewChildKey(Purpose)
	if err != nil {
		log.Println("Purpose error: ", err.Error())
		return
	} 

	child, err = child.NewChildKey(CoinEther)
	if err != nil {
		log.Println("CoinEther error: ", err.Error())
		return
	}

	child, err = child.NewChildKey(Account)
	if err != nil {
		log.Println("Account error: ", err.Error())
		return
	}	

	childAccounts, _ := child.NewChildKey(SubAccounts)
	if err != nil {
		log.Println("External error: ", err.Error())
		return
	}

	childPrivateKey, _ := childAccounts.NewChildKey(uint32(level))
	hexKey := hexutil.Encode(childPrivateKey.Key)[2:]
	privateKey, _ = crypto.HexToECDSA(hexKey)
	fromAddress = crypto.PubkeyToAddress(privateKey.PublicKey)

	return
	
}

//EthAccountTransfer ...
func EthAccountTransfer(amount float64, fromAddress, toAddress common.Address, privateKey *ecdsa.PrivateKey) {
	//Get the Nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(int64(amount * WEI)) // in WEI (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: https://rinkeby.etherscan.io/tx/%s \n", signedTx.Hash().Hex())

}

//EthAccountBal ...
func EthAccountBal(address string, block int64) (balance *big.Int, err error) {
	if client == nil {
		EthClientDial()
	}

	if client != nil {
		var blockNumber *big.Int
		if block > 0 {
			blockNumber = big.NewInt(block)
		}
		account := common.HexToAddress(address)
		balance, err = client.BalanceAt(context.Background(), account, blockNumber)

		if err != nil {
			log.Println("Account Balannce error: ", err.Error())
		}
	}
	return
}

func ETHAccountBalFloat(address string, block int64) (*big.Float, error) {
	bal, err := EthAccountBal(address, block )
	balFloat, weiFloat := new(big.Float).SetInt(bal), big.NewFloat(WEI)
	newBalance := new(big.Float).Quo(balFloat, weiFloat)
	return newBalance, err
}
