package pancake_swap

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"

	"github.com/medasx/turtle-snipe/pancake-swap/contract"
	"github.com/medasx/turtle-snipe/pancake-swap/router"
)

const speedyNodeURL = "wss://speedy-nodes-nyc.moralis.io/7dea25b6c058d4f577fc37c4/bsc/mainnet/ws"

func NewClient(ctx context.Context, privateKey string) (*Client, error) {
	client, err := ethclient.DialContext(ctx, speedyNodeURL)
	if err != nil {
		return nil, err
	}
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("privateKey : %w", err)
	}
	publicKey := pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid key")
	}
	userAddr := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("ChainID: %w", err)
	}
	return &Client{
		client:      client,
		privateKey:  pk,
		userAddress: userAddr,
		ChainID:     chainID,
	}, nil
}

type Client struct {
	client      *ethclient.Client
	privateKey  *ecdsa.PrivateKey
	userAddress common.Address

	ChainID *big.Int
}

func (c *Client) ListenToPairCreated(ctx context.Context, contractAddr string) (event.Subscription, <-chan *contract.IPancakeFactoryPairCreated, error) {
	contractAddress := common.HexToAddress(contractAddr)
	factoryFilterer, err := contract.NewIPancakeFactoryFilterer(contractAddress, c.client)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create IPancakeFactoryPairFilterer: %w", err)
	}
	s := make(chan *contract.IPancakeFactoryPairCreated)

	sub, err := factoryFilterer.WatchPairCreated(&bind.WatchOpts{Context: ctx}, s, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("could not subcribe to PairCreated: %w", err)
	}
	return sub, s, nil
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) Buy(ctx context.Context, contractAddress common.Address, amount *big.Int) error {
	fmt.Printf("Buying for %s BNB, contract: %s\n", WeiToEth(amount).String(), contractAddress)
	nonce, err := c.client.PendingNonceAt(context.Background(), c.userAddress)
	if err != nil {
		return fmt.Errorf("PendingNonceAt: %w", err)
	}
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		fmt.Println(fmt.Errorf("SuggestGasPrice: %w", err))
		gasPrice = big.NewInt(1000000000)
	}

	fmt.Printf("Gas price: %s\n", WeiToEth(gasPrice).String())
	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.ChainID)
	if err != nil {
		return fmt.Errorf("NewKeyedTransactorWithChainID: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = amount
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice
	auth.Context = ctx
	r, err := router.NewIPancakeRouter02(RouterAddr, c.client)
	if err != nil {
		return fmt.Errorf("NewIPancakeRouter02: %w", err)
	}
	path := []common.Address{WBNBAddr, contractAddress}
	amountsOut, err := r.GetAmountsOut(
		&bind.CallOpts{From: c.userAddress, Context: ctx},
		amount,
		path,
	)
	if err != nil {
		return fmt.Errorf("GetAmountOut: %w", err)
	}
	for i := range amountsOut {
		fmt.Printf("\tamount out: %s\n", WeiToEth(amountsOut[i]).String())
	}
	tx, err := r.SwapExactETHForTokensSupportingFeeOnTransferTokens(
		auth,
		new(big.Int).Div(amountsOut[len(path)-1], big.NewInt(2)),
		path,
		c.userAddress,
		big.NewInt(time.Now().Add(time.Minute*20).Unix()),
	)
	if err != nil {
		return fmt.Errorf("SwapExactETHForTokensSupportingFeeOnTransferTokens: %w", err)
	}

	fmt.Printf("Pending TX: 0x%x\n", tx.Hash())

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	var ch = make(chan types.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, query, ch)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

outer:
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-ch:
			if vLog.TxHash.Hex() == tx.Hash().Hex() {
				break outer
			}
		}

	}

	receipt, err := c.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("TransactionReceipt: %w", err)
	}
	switch receipt.Status {
	case types.ReceiptStatusFailed:
		fmt.Println("failed")
	case types.ReceiptStatusSuccessful:
		fmt.Println("succeed")
	}

	fmt.Printf("Receipt %+v\n", *receipt)
	token, err := contract.NewIERC20(contractAddress, c.client)
	if err != nil {
		return err
	}
	tname, err := token.Name(&bind.CallOpts{Context: ctx})
	if err != nil {
		return err
	}
	balance, err := token.BalanceOf(&bind.CallOpts{Context: ctx}, c.userAddress)
	if err != nil {
		return err
	}
	fmt.Printf("Balance: %s %s\n", WeiToEth(balance).String(), tname)
	return nil
}
