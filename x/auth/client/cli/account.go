package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/QOSGroup/qstars/client/context"
	sdk "github.com/QOSGroup/qstars/types"
	"github.com/QOSGroup/qstars/wire"
	"github.com/QOSGroup/qstars/x/auth"
	qos "github.com/QOSGroup/qos/account"
	qstarstypes "github.com/QOSGroup/qstars/types"
)



//GetAccountDecoder gets the account decoder for auth.DefaultAccount.
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) ( auth.QAccount,  error) {
		qacc := qos.QOSAccount{}
		var err = cdc.UnmarshalBinary(accBytes, &qacc)
		if err != nil {
			panic(err)
		}

		//var coins [len(qacc.QscList)]qstarstypes.Coin
		var coins qstarstypes.Coins
		for _, qsc := range qacc.QscList{
			amount := qsc.Amount
			coins = append(coins, qstarstypes.NewCoin(qsc.Name,qstarstypes.NewInt(amount.Int64())))
		}
		acct := auth.QStarsAccount{QosAccount:qacc,QCoins:coins}

		return acct, err
	}
}

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
//, decoder auth.AccountDecoder
func GetAccountCmd(storeName string, cdc *wire.Codec,decoder auth.AccountDecoder) *cobra.Command {
	return &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// find the key to look up the account
			addr := args[0]

			key, err := sdk.AccAddressFromBech32(addr)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(decoder)

			// in qstars, we don't need to ensure it
			//if err := cliCtx.EnsureAccountExistsFromAddr(key); err != nil {
			//	return err
			//}

			acc, err := cliCtx.GetAccount(key)
			if err != nil {
				return err
			}

			output, err := wire.MarshalJSONIndent(cdc, acc)
			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}
}