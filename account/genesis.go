package account

import (
	"code.aliyun.com/chain33/chain33/types"
)

func (acc *AccountDB) GenesisInit(addr string, amount int64) (*types.Receipt, error) {
	accTo := acc.LoadAccount(addr)
	copyto := *accTo
	accTo.Balance = accTo.GetBalance() + amount
	receiptBalanceTo := &types.ReceiptAccountTransfer{&copyto, accTo}
	acc.SaveAccount(accTo)
	receipt := acc.genesisReceipt(accTo, receiptBalanceTo)
	return receipt, nil
}

func (acc *AccountDB) GenesisInitExec(addr string, amount int64, execaddr string) (*types.Receipt, error) {
	accTo := acc.LoadAccount(execaddr)
	copyto := *accTo
	accTo.Balance = accTo.GetBalance() + amount
	receiptBalanceTo := &types.ReceiptAccountTransfer{&copyto, accTo}
	acc.SaveAccount(accTo)
	receipt := acc.genesisReceipt(accTo, receiptBalanceTo)
	receipt2, err := acc.execDeposit(addr, execaddr, amount)
	if err != nil {
		panic(err)
	}
	receipt2.Ty = types.TyLogGenesisDeposit
	receipt = acc.mergeReceipt(receipt, receipt2)
	return receipt, nil
}

func (acc *AccountDB) genesisReceipt(accTo *types.Account, receiptTo *types.ReceiptAccountTransfer) *types.Receipt {
	log2 := &types.ReceiptLog{types.TyLogGenesisTransfer, types.Encode(receiptTo)}
	kv := acc.GetKVSet(accTo)
	return &types.Receipt{types.ExecOk, kv, []*types.ReceiptLog{log2}}
}