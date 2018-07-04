package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"gitlab.33.cn/chain33/chain33/common"
	jsonrpc "gitlab.33.cn/chain33/chain33/rpc"
	"gitlab.33.cn/chain33/chain33/types"
)

var (
	defMixCount int32 = 16
)

func PrivacyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "privacy",
		Short: "Privacy transaction management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		ShowPrivacyKeyCmd(),
		ShowPrivacyAccountSpendCmd(),
		Public2PrivacyCmd(),
		Privacy2PrivacyCmd(),
		Privacy2PublicCmd(),
		ShowAmountsOfUTXOCmd(),
		ShowUTXOs4SpecifiedAmountCmd(),
		CreateUTXOsCmd(),
		ShowPrivacyAccountInfoCmd(),
	)

	return cmd
}

// ShowPrivacyKeyCmd show privacy key by address
func ShowPrivacyKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpk",
		Short: "Show privacy key by address",
		Run:   showPrivacyKey,
	}
	showPrivacyKeyFlag(cmd)
	return cmd
}

func showPrivacyKeyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")

}

func showPrivacyKey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := types.ReqStr{
		ReqStr: addr,
	}

	var res types.ReplyPrivacyPkPair
	ctx := NewRpcCtx(rpcLaddr, "Chain33.ShowPrivacykey", params, &res)
	ctx.Run()
}

// Public2PrivacyCmd public address to privacy address
func Public2PrivacyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pub2priv",
		Short: "Public to privacy from pubkeypair amout note",
		Run:   public2Privacy,
	}
	public2PrivacyFlag(cmd)
	return cmd
}

func public2PrivacyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "from account address")
	cmd.MarkFlagRequired("from")

	cmd.Flags().StringP("pubkeypair", "p", "", "to view spend public key pair")
	cmd.MarkFlagRequired("pubkeypair")

	cmd.Flags().Float64P("amount", "a", 0, "transfer amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transfer note")
	cmd.MarkFlagRequired("note")

}

func public2Privacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	pubkeypair, _ := cmd.Flags().GetString("pubkeypair")
	amount, _ := cmd.Flags().GetFloat64("amount")
	note, _ := cmd.Flags().GetString("note")

	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4 //支持4位小数输入，多余的输入将被截断
	params := types.ReqPub2Pri{
		Sender:     from,
		Pubkeypair: pubkeypair,
		Amount:     amountInt64,
		Note:       note,
		Tokenname:  types.BTY,
		Expire:     int64(time.Hour),
	}

	var res jsonrpc.ReplyHash
	ctx := NewRpcCtx(rpcLaddr, "Chain33.MakeTxPublic2privacy", params, &res)
	ctx.Run()
}

// privacy address to privacy address
func Privacy2PrivacyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priv2priv",
		Short: "Privacy to privacy from toviewpubkey tospendpubkey amout note",
		Run:   privacy2Privacy,
	}
	privacy2PrivacyFlag(cmd)
	return cmd
}

func privacy2PrivacyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "from account address")
	cmd.MarkFlagRequired("from")

	cmd.Flags().StringP("pubkeypair", "p", "", "to view spend public key pair")
	cmd.MarkFlagRequired("pubkeypair")

	cmd.Flags().Float64P("amount", "a", 0.0, "transfer amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transfer note")
	cmd.MarkFlagRequired("note")

	cmd.Flags().Int32P("mixcount", "m", defMixCount, "transfer note")
}

func privacy2Privacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	pubkeypair, _ := cmd.Flags().GetString("pubkeypair")
	amount, _ := cmd.Flags().GetFloat64("amount")
	mixcount, _ := cmd.Flags().GetInt32("mixcount")
	note, _ := cmd.Flags().GetString("note")

	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4 //支持4位小数输入，多余的输入将被截断
	params := types.ReqPri2Pri{
		Sender:     from,
		Pubkeypair: pubkeypair,
		Amount:     amountInt64,
		Mixin:      mixcount,
		Note:       note,
		Tokenname:  types.BTY,
		Expire:     int64(time.Hour),
	}

	var res jsonrpc.ReplyHash
	ctx := NewRpcCtx(rpcLaddr, "Chain33.MakeTxPrivacy2privacy", params, &res)
	ctx.Run()
}

// privacy address to public address
func Privacy2PublicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priv2pub",
		Short: "Public to privacy from toviewpubkey tospendpubkey amout note",
		Run:   privacy2Public,
	}
	privacy2Publiclag(cmd)
	return cmd
}

func privacy2Publiclag(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "from account address")
	cmd.MarkFlagRequired("from")

	cmd.Flags().StringP("to", "t", "", "to account address")
	cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0.0, "transfer amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transfer note")
	cmd.MarkFlagRequired("note")

	cmd.Flags().Int32P("mixcount", "m", defMixCount, "transfer note")

}

func privacy2Public(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetFloat64("amount")
	mixcount, _ := cmd.Flags().GetInt32("mixcount")
	note, _ := cmd.Flags().GetString("note")

	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4 //支持4位小数输入，多余的输入将被截断
	params := types.ReqPri2Pub{
		Sender:    from,
		Receiver:  to,
		Amount:    amountInt64,
		Note:      note,
		Tokenname: types.BTY,
		Mixin:     mixcount,
		Expire:    int64(time.Hour),
	}

	var res jsonrpc.ReplyHash
	ctx := NewRpcCtx(rpcLaddr, "Chain33.MakeTxPrivacy2public", params, &res)
	ctx.Run()
}

func ShowPrivacyAccountSpendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpas",
		Short: "Show privacy account spend command",
		Run:   showPrivacyAccountSpend,
	}
	showPrivacyAccountSpendFlag(cmd)
	return cmd
}

func showPrivacyAccountSpendFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")
}

func showPrivacyAccountSpend(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := types.ReqPrivBal4AddrToken{
		Addr:  addr,
		Token: types.BTY,
	}

	var res types.UTXOHaveTxHashs
	ctx := NewRpcCtx(rpcLaddr, "Chain33.ShowPrivacyAccountSpend", params, &res)
	ctx.SetResultCb(parseShowPrivacyAccountSpendRes)
	ctx.Run()
}

func parseShowPrivacyAccountSpendRes(arg interface{}) (interface{}, error) {
	total := float64(0)
	res := arg.(*types.UTXOHaveTxHashs)
	rets := make([]*PrivacyAccountSpendResult, 0)
	for _, utxo := range res.UtxoHaveTxHashs {
		amount := float64(utxo.Amount) / float64(types.Coin)
		total += amount

		var isSave bool
		for _, ret := range rets {
			if utxo.TxHash == ret.Txhash {
				result := &PrivacyAccountResult{
					Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
					OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
					Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
				}
				ret.Res = append(ret.Res, result)
				isSave = true
				break
			}
		}

		if !isSave {
			result := &PrivacyAccountResult{
				//Height:   utxo.UtxoBasic.UtxoGlobalIndex.Height,
				//TxIndex:  utxo.UtxoBasic.UtxoGlobalIndex.Txindex,
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			var SpendResult PrivacyAccountSpendResult
			SpendResult.Txhash = utxo.TxHash
			SpendResult.Res = append(SpendResult.Res, result)
			rets = append(rets, &SpendResult)
		}
	}

	fmt.Println(fmt.Sprintf("Total Privacy spend amount is %s", strconv.FormatFloat(total, 'f', 4, 64)))

	return rets, nil
}

func ShowAmountsOfUTXOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showau",
		Short: "Show Amount of UTXO",
		Run:   showAmountOfUTXO,
	}
	showAmountOfUTXOFlag(cmd)
	return cmd
}

func showAmountOfUTXOFlag(cmd *cobra.Command) {
}

func showAmountOfUTXO(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	reqPrivacyToken := types.ReqPrivacyToken{Token: types.BTY}
	var params jsonrpc.Query4Cli
	params.Execer = types.PrivacyX
	params.FuncName = "ShowAmountsOfUTXO"
	params.Payload = reqPrivacyToken

	var res types.ReplyPrivacyAmounts
	ctx := NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(parseShowAmountOfUTXORes)
	ctx.Run()
}

func parseShowAmountOfUTXORes(arg interface{}) (interface{}, error) {
	res := arg.(*types.ReplyPrivacyAmounts)
	for _, amount := range res.AmountDetail {
		amount.Amount = amount.Amount / types.Coin
	}
	return res, nil
}

func ShowUTXOs4SpecifiedAmountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showutxo4a",
		Short: "Show specified amount UTXOs",
		Run:   showUTXOs4SpecifiedAmount,
	}
	showUTXOs4SpecifiedAmountFlag(cmd)
	return cmd
}

func showUTXOs4SpecifiedAmountFlag(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")
}

func showUTXOs4SpecifiedAmount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	amount, _ := cmd.Flags().GetFloat64("amount")
	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4

	reqPrivacyToken := types.ReqPrivacyToken{
		Token:  types.BTY,
		Amount: amountInt64,
	}
	var params jsonrpc.Query4Cli
	params.Execer = types.PrivacyX
	params.FuncName = "ShowUTXOs4SpecifiedAmount"
	params.Payload = reqPrivacyToken

	var res types.ReplyUTXOsOfAmount
	ctx := NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(parseShowUTXOs4SpecifiedAmountRes)
	ctx.Run()
}

func parseShowUTXOs4SpecifiedAmountRes(arg interface{}) (interface{}, error) {
	res := arg.(*types.ReplyUTXOsOfAmount)
	ret := make([]*PrivacyAccountResult, 0)
	for _, item := range res.LocalUTXOItems {
		result := &PrivacyAccountResult{
			Txhash:        common.ToHex(item.Txhash),
			OutIndex:      item.Outindex,
			OnetimePubKey: common.Bytes2Hex(item.Onetimepubkey),
		}
		ret = append(ret, result)
	}

	return ret, nil
}

func CreateUTXOsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createutxos",
		Short: "Create specified count UTXOs with specified amount",
		Run:   createUTXOs,
	}
	createUTXOsFlag(cmd)
	return cmd
}

func createUTXOsFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "from account address")
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("pubkeypair", "p", "", "to view spend public key pair")
	cmd.MarkFlagRequired("pubkeypair")
	cmd.Flags().Float64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Int32P("count", "c", 0, "mix count")
	cmd.MarkFlagRequired("count")
	cmd.Flags().StringP("note", "n", "", "transfer note")
	cmd.MarkFlagRequired("note")
}

func createUTXOs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	pubkeypair, _ := cmd.Flags().GetString("pubkeypair")
	note, _ := cmd.Flags().GetString("note")
	count, _ := cmd.Flags().GetInt32("count")
	amount, _ := cmd.Flags().GetFloat64("amount")
	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4

	params := &types.ReqCreateUTXOs{
		Tokenname:  types.BTY,
		Sender:     from,
		Pubkeypair: pubkeypair,
		Amount:     amountInt64,
		Count:      count,
		Note:       note,
		Expire:     int64(time.Hour),
	}

	var res jsonrpc.ReplyHash
	ctx := NewRpcCtx(rpcLaddr, "Chain33.CreateUTXOs", params, &res)
	ctx.Run()
}

// ShowPrivacyAccountInfoCmd
func ShowPrivacyAccountInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpai",
		Short: "Show privacy account information",
		Run:   showPrivacyAccountInfo,
	}
	showPrivacyAccountInfoFlag(cmd)
	return cmd
}

func showPrivacyAccountInfoFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("token", "t", types.BTY, "coins token, BTY supported.")
	cmd.Flags().Int32P("displaymode", "d", 0, "display mode.(0: display collect. 1:display avaliable detail. 2:display frozen detail. 3:display all")
}

func showPrivacyAccountInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	token, _ := cmd.Flags().GetString("token")
	mode, _ := cmd.Flags().GetInt32("displaymode")
	if mode < 0 || mode > 3 {
		fmt.Println("display mode only support 0-3")
		return
	}

	params := types.ReqPPrivacyAccount{
		Addr:        addr,
		Token:       token,
		Displaymode: mode,
	}

	var res types.ReplyPrivacyAccount
	ctx := NewRpcCtx(rpcLaddr, "Chain33.ShowPrivacyAccountInfo", params, &res)
	ctx.SetResultCb(parseshowPrivacyAccountInfo)
	ctx.Run()
}

func parseshowPrivacyAccountInfo(arg interface{}) (interface{}, error) {
	total := float64(0)
	totalFrozen := float64(0)
	res := arg.(*types.ReplyPrivacyAccount)

	var availableAmount, frozenAmount, totalAmount string

	utxos := make([]*PrivacyAccountResult, 0)
	for _, utxo := range res.Utxos.Utxos {
		amount := float64(utxo.Amount) / float64(types.Coin)
		total += amount

		if res.Displaymode == 1 || res.Displaymode == 3 {
			result := &PrivacyAccountResult{
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			utxos = append(utxos, result)
		}
	}
	availableAmount = strconv.FormatFloat(total, 'f', 4, 64)

	ftxos := make([]*PrivacyAccountResult, 0)
	for _, utxo := range res.Ftxos.Utxos {
		amount := float64(utxo.Amount) / float64(types.Coin)
		totalFrozen += amount

		if res.Displaymode == 2 || res.Displaymode == 3 {
			result := &PrivacyAccountResult{
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			ftxos = append(ftxos, result)
		}
	}
	frozenAmount = strconv.FormatFloat(totalFrozen, 'f', 4, 64)
	totalAmount = strconv.FormatFloat(total+totalFrozen, 'f', 4, 64)

	ret := &PrivacyAccountInfoResult{
		AvailableDetail: utxos,
		FrozenDetail:    ftxos,
		AvailableAmount: availableAmount,
		FrozenAmount:    frozenAmount,
		TotalAmount:     totalAmount,
	}
	return ret, nil
}