package main

import (
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	pancake "github.com/medasx/turtle-snipe/pancake-swap"
)

func main() {
	var (
		client       *pancake.Client
		pk           string
		contractAddr string
		amount       float64
	)
	app := &cli.App{
		Name:  "snipe",
		Usage: "buy tokens right after creating liquidity pair",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "private-key",
				Aliases:     []string{"pk"},
				Required:    true,
				Destination: &pk,
				Usage:       "wallet's private key",
			},
		},
		Before: func(ctx *cli.Context) error {
			c, err := pancake.NewClient(ctx.Context, pk)
			if err != nil {
				return err
			}
			client = c
			return nil
		},
		After: func(_ *cli.Context) error {
			client.Close()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "buy",
				Aliases: []string{"b"},
				Usage:   "buy token for bnb",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Destination: &contractAddr,
						Aliases:     []string{"c"},
						Required:    true,
						Name:        "contract",
						Usage:       "contract address",
					},
					&cli.Float64Flag{
						Name:        "amount",
						Required:    true,
						Destination: &amount,
					},
				},
				Action: func(ctx *cli.Context) error {
					return client.Buy(ctx.Context, common.HexToAddress(contractAddr), pancake.EthToWei(big.NewFloat(amount)))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
