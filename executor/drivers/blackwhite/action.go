package blackwhite

import (
	"bytes"
	"gitlab.33.cn/chain33/chain33/account"
	"gitlab.33.cn/chain33/chain33/common"
	dbm "gitlab.33.cn/chain33/chain33/common/db"
	"gitlab.33.cn/chain33/chain33/types"
	gt "gitlab.33.cn/chain33/chain33/types/executor/blackwhite"
	"math"
	"time"
)

const (
	MaxAmount      int64 = 20 * types.Coin
	MinAmount      int64 = 1 * types.Coin
	MinPlayerCount int32 = 3
	MaxPlayerCount int32 = 100000
	lockAmount     int64 = types.Coin / 100        //创建者锁定金额
	showTimeout    int64 = int64(time.Minute) * 5  // 公布密钥超时时间
	MaxPlayTimeout int64 = int64(time.Hour) * 24   // 创建交易之后最大超时时间
	MinPlayTimeout int64 = int64(time.Minute) * 10 // 创建交易之后最小超时时间

	white = 0
	black = 1
)

type action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
}

type resultCalc struct {
	Addr    string
	amount  int64
	IsWin   bool
	IsBlack []bool
}

type addrResult struct {
	addr   string
	amount int64
}

func newAction(t *Blackwhite, tx *types.Transaction) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{t.GetCoinsAccount(), t.GetStateDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), t.GetAddr()}
}

func (a *action) Create(create *types.BlackwhiteCreate) (*types.Receipt, error) {
	if create.PlayAmount < MinAmount || create.PlayAmount > MaxAmount {
		return nil, types.ErrInputPara
	}
	if create.PlayerCount < MinPlayerCount || create.PlayerCount > MaxPlayerCount {
		return nil, types.ErrInputPara
	}
	if create.Timeout < MinPlayTimeout || create.Timeout > MaxPlayTimeout {
		return nil, types.ErrInputPara
	}

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, lockAmount)
	if err != nil {
		clog.Error("blackwhite create ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", lockAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	round := newRound(create, a.fromaddr)
	round.GameID = common.ToHex(a.txhash)
	round.CreateTime = a.blocktime

	key := calcRoundKey(round.GameID)
	value := types.Encode(round)

	kv = append(kv, &types.KeyValue{key, value})

	//log := &types.ReceiptBlackwhite{round}
	//logs = append(logs, &types.ReceiptLog{types.TyLogBlackwhiteCreate, types.Encode(log)})

	receiptLog := a.GetReceiptLog(round)
	logs = append(logs, receiptLog)

	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

func (a *action) Play(play *types.BlackwhitePlay) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(calcRoundKey(play.GameID))
	if err != nil {
		clog.Error("blackwhite play ", "addr", a.fromaddr, "execaddr", a.execaddr, "get round failed",
			play.GameID, "err", err)
		return nil, err
	}
	var round types.BlackwhiteRound
	err = types.Decode(value, &round)
	if err != nil {
		clog.Error("blackwhite play ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode round failed",
			play.GameID, "err", err)
		return nil, err
	}

	// 检查当前状态
	if gt.BlackwhiteStatusPlay != round.Status && gt.BlackwhiteStatusCreate != round.Status {
		err := types.ErrGameOver
		clog.Error("blackwhite play ", "addr", a.fromaddr, "round status", round.Status, "status is not match GameID ",
			play.GameID, "err", err)
		return nil, err
	}

	// 检查是否有重复
	for _, addrResult := range round.AddrResult {
		if addrResult.Addr == a.fromaddr {
			err := types.ErrOnceRoundRepeatPlay
			clog.Error("blackwhite play ", "addr", a.fromaddr, "execaddr", a.execaddr, "repeat address GameID",
				play.GameID, "err", err)
			return nil, err
		}
	}

	if play.Amount < round.PlayAmount {
		clog.Error("blackwhite play ", "addr", a.fromaddr, "execaddr", a.execaddr, "playAmount < roundAmount in this GameID ",
			play.GameID)
		return nil, types.ErrInputPara
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, play.Amount)
	if err != nil {
		clog.Error("blackwhite Play ", "addr", a.fromaddr, "execaddr", a.execaddr, "ExecFrozen amount", play.Amount)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	round.Status = gt.BlackwhiteStatusPlay
	addrRes := &types.AddressResult{
		Addr:       a.fromaddr,
		Amount:     play.Amount,
		HashValues: play.HashValues,
	}
	round.AddrResult = append(round.AddrResult, addrRes)
	round.CurPlayerCount++

	//tyLog := 0
	if round.CurPlayerCount >= round.PlayerCount {
		// 触发进入到公布阶段
		round.ShowTime = a.blocktime
		round.Status = gt.BlackwhiteStatusShow
		//tyLog = types.TyLogBlackwhiteShow
	} else {
		//tyLog = types.TyLogBlackwhitePlay
	}

	key1 := calcRoundKey(round.GameID)
	value1 := types.Encode(&round)
	kv = append(kv, &types.KeyValue{key1, value1})

	//log := &types.ReceiptBlackwhite{&round}
	//logs = append(logs, &types.ReceiptLog{int32(tyLog), types.Encode(log)})

	receiptLog := a.GetReceiptLog(&round)
	logs = append(logs, receiptLog)

	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

func (a *action) Show(show *types.BlackwhiteShow) (*types.Receipt, error) {
	// 获取GameID
	value, err := a.db.Get(calcRoundKey(show.GameID))
	if err != nil {
		clog.Error("blackwhite show ", "addr", a.fromaddr, "execaddr", a.execaddr, "get round failed",
			show.GameID, "err", err)
		return nil, err
	}
	var round types.BlackwhiteRound
	err = types.Decode(value, &round)
	if err != nil {
		clog.Error("blackwhite show ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode round failed",
			show.GameID, "err", err)
		return nil, err
	}
	// 检查当前状态
	if gt.BlackwhiteStatusShow != round.Status {
		err := types.ErrNoShowNonce
		clog.Error("blackwhite show ", "addr", a.fromaddr, "round status", round.Status, "status is not match GameID ",
			show.GameID, "err", err)
		return nil, err
	}

	// 检查是否存在该地址押金
	bIsExist := false
	index := 0
	for i, addrResult := range round.AddrResult {
		if addrResult.Addr == a.fromaddr {
			bIsExist = true
			index = i
			break
		}
	}
	if false == bIsExist {
		err := types.ErrNoExistAddr
		clog.Error("blackwhite show ", "addr", a.fromaddr, "execaddr", a.execaddr, "this addr is play in GameID",
			show.GameID, "err", err)
		return nil, err
	}
	//更新信息
	round.Status = gt.BlackwhiteStatusShow
	round.AddrResult[index].ShowSecret = show.Secret
	round.CurShowCount++

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//tyLog := 0

	if round.CurShowCount >= round.PlayerCount {
		// 已经集齐有所有密钥
		round.Status = gt.BlackwhiteStatusDone
		receipt, err := a.StatTransfer(&round)
		if err != nil {
			clog.Error("blackwhite show fail", "StatTransfer err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		//tyLog = types.TyLogBlackwhiteDone
	} else {
		//tyLog = types.TyLogBlackwhiteShow
	}

	key1 := calcRoundKey(round.GameID)
	value1 := types.Encode(&round)
	kv = append(kv, &types.KeyValue{key1, value1})

	//log := &types.ReceiptBlackwhite{&round}
	//logs = append(logs, &types.ReceiptLog{int32(tyLog), types.Encode(log)})

	receiptLog := a.GetReceiptLog(&round)
	logs = append(logs, receiptLog)

	return &types.Receipt{types.ExecOk, kv, logs}, nil
}

func (a *action) TimeoutDone(done *types.BlackwhiteTimeoutDone) (*types.Receipt, error) {
	value, err := a.db.Get(calcRoundKey(done.GameID))
	if err != nil {
		clog.Error("blackwhite timeout done ", "addr", a.fromaddr, "execaddr", a.execaddr, "get round failed",
			done.GameID, "err", err)
		return nil, err
	}

	var round types.BlackwhiteRound
	err = types.Decode(value, &round)
	if err != nil {
		clog.Error("blackwhite timeout done ", "addr", a.fromaddr, "execaddr", a.execaddr, "decode round failed",
			done.GameID, "err", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//tyLog := 0

	// 检查当前状态
	if gt.BlackwhiteStatusPlay == round.Status {
		if a.blocktime >= round.Timeout+round.CreateTime {
			//进行超时play超时处理，即将所有冻结资金都解冻，然后游戏结束
			for i, addrRes := range round.AddrResult {
				receipt, err := a.coinsAccount.ExecActive(addrRes.Addr, a.execaddr, addrRes.Amount)
				if err != nil {
					//rollback
					for j, addrR := range round.AddrResult {
						if j < i {
							a.coinsAccount.ExecFrozen(addrR.Addr, a.execaddr, addrR.Amount)
						} else {
							break
						}
					}
					clog.Error("blackwhite timeout done", "addr", a.fromaddr, "execaddr", a.execaddr, "execActive all player GameID", done.GameID, "err", err)
					return nil, err
				}
				logs = append(logs, receipt.Logs...)
				kv = append(kv, receipt.KV...)
			}
			// 将创建游戏者解冻
			receipt, err := a.coinsAccount.ExecActive(round.CreateAddr, a.execaddr, lockAmount)
			if err != nil {
				for _, addrR := range round.AddrResult {
					a.coinsAccount.ExecFrozen(addrR.Addr, a.execaddr, addrR.Amount)
				}
				clog.Error("blackwhite timeout done", "addr", round.CreateAddr, "execaddr", a.execaddr, "execActive create lockAmount", lockAmount, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)

			round.Status = gt.BlackwhiteStatusTimeout

		} else {
			err := types.ErrNoTimeoutDone
			clog.Error("blackwhite timeout done ", "addr", a.fromaddr, "execaddr", a.execaddr, "is BlackwhiteStatusPlay GameID",
				done.GameID, "err", err)
			return nil, err
		}
	} else if gt.BlackwhiteStatusShow == round.Status {
		if a.blocktime >= showTimeout+round.ShowTime {
			//show私钥超时,有私钥的进行开奖
			round.Status = gt.BlackwhiteStatusDone
			receipt, err := a.StatTransfer(&round)
			if err != nil {
				clog.Error("blackwhite done fail", "StatTransfer err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
			//tyLog = types.TyLogBlackwhiteDone

		} else {
			err := types.ErrNoTimeoutDone
			clog.Error("blackwhite timeout done ", "addr", a.fromaddr, "execaddr", a.execaddr, "is blackwhiteStatusShow GameID",
				done.GameID, "err", err)
			return nil, err
		}
	} else {
		err := types.ErrNoTimeoutDone
		clog.Error("blackwhite timeout done ", "addr", a.fromaddr, "execaddr", a.execaddr, "status is not match GameID",
			done.GameID, "status", round.Status, "err", err)
		return nil, err
	}

	key1 := calcRoundKey(round.GameID)
	value1 := types.Encode(&round)
	kv = append(kv, &types.KeyValue{key1, value1})

	//log := &types.ReceiptBlackwhite{&round}
	//logs = append(logs, &types.ReceiptLog{int32(tyLog), types.Encode(log)})

	receiptLog := a.GetReceiptLog(&round)
	logs = append(logs, receiptLog)

	return &types.Receipt{types.ExecOk, kv, logs}, nil

}

func (a *action) StatTransfer(round *types.BlackwhiteRound) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	winers, loopResults := a.getWinner(round)
	Losers := a.getLoser(round)
	var averAmount int64

	if len(winers) == 0 {
		// 将所有参与人员都解冻
		for i, addrRes := range round.AddrResult {
			receipt, err := a.coinsAccount.ExecActive(addrRes.Addr, a.execaddr, addrRes.Amount)
			if err != nil {
				//rollback
				for j, addrR := range round.AddrResult {
					if j < i {
						a.coinsAccount.ExecFrozen(addrR.Addr, a.execaddr, addrR.Amount)
					} else {
						break
					}
				}
				clog.Error("StatTransfer execActive no winers", "addr", a.fromaddr, "execaddr", a.execaddr, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	} else {

		var sumAmount int64
		for i, Loser := range Losers {
			// 将其转入黑白配合约的合约地址
			sumAmount += Loser.amount
			receipt, err := a.coinsAccount.ExecTransferFrozen(Loser.addr, blackwhiteAddr, a.execaddr, Loser.amount)
			if err != nil {
				//rollback
				for j, addrR := range Losers {
					if j < i {
						a.coinsAccount.ExecTransfer(blackwhiteAddr, addrR.addr, a.execaddr, addrR.amount)
						a.coinsAccount.ExecFrozen(addrR.addr, a.execaddr, addrR.amount)
					} else {
						break
					}
				}
				clog.Error("StatTransfer all losers to blackwhiteAddr", "addr", a.fromaddr, "execaddr", a.execaddr, "amount", Loser.amount)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		var winNum int64
		winNum = int64(len(winers))
		averAmount = sumAmount / winNum
		// 从公共账户转帐给它获胜用户
		for i, winer := range winers {
			receipt, err := a.coinsAccount.ExecTransfer(blackwhiteAddr, winer.addr, a.execaddr, averAmount)
			if err != nil {
				//rollback
				for j, winer := range winers {
					if j < i {
						a.coinsAccount.ExecTransfer(winer.addr, blackwhiteAddr, a.execaddr, averAmount)
					} else {
						break
					}
				}
				for _, loser := range Losers {
					a.coinsAccount.ExecTransfer(blackwhiteAddr, loser.addr, a.execaddr, loser.amount)
					a.coinsAccount.ExecFrozen(loser.addr, a.execaddr, loser.amount)
				}
				clog.Error("StatTransfer one winer to any other winers fail", "addr", winer, "execaddr", a.execaddr, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		// 胜利人员都解冻
		for i, winer := range winers {
			receipt, err := a.coinsAccount.ExecActive(winer.addr, a.execaddr, winer.amount)
			if err != nil {
				//rollback
				for j, winer := range winers {
					if j < i {
						a.coinsAccount.ExecFrozen(winer.addr, a.execaddr, winer.amount)
					} else {
						break
					}
				}
				for _, winer := range winers {
					a.coinsAccount.ExecTransfer(winer.addr, blackwhiteAddr, a.execaddr, averAmount)
				}
				for _, loser := range Losers {
					a.coinsAccount.ExecTransfer(blackwhiteAddr, loser.addr, a.execaddr, loser.amount)
					a.coinsAccount.ExecFrozen(loser.addr, a.execaddr, loser.amount)
				}
				clog.Error("StatTransfer ExecActive have winers", "addr", a.fromaddr, "execaddr", a.execaddr, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}

	for _, winer := range winers {
		round.Winner = append(round.Winner, winer.addr)
	}

	// 将创建游戏者解冻
	receipt, err := a.coinsAccount.ExecActive(round.CreateAddr, a.execaddr, lockAmount)
	if err != nil {
		// rollback
		if len(winers) == 0 {
			for _, addrR := range round.AddrResult {
				a.coinsAccount.ExecFrozen(addrR.Addr, a.execaddr, addrR.Amount)
			}
		} else {
			for _, winer := range winers {
				a.coinsAccount.ExecFrozen(winer.addr, a.execaddr, winer.amount)
			}
			for _, winer := range winers {
				a.coinsAccount.ExecTransfer(winer.addr, blackwhiteAddr, a.execaddr, averAmount)
			}
			for _, loser := range Losers {
				a.coinsAccount.ExecTransfer(blackwhiteAddr, loser.addr, a.execaddr, loser.amount)
				a.coinsAccount.ExecFrozen(loser.addr, a.execaddr, loser.amount)
			}
		}
		clog.Error("StatTransfer ExecActive create ExecFrozen ", "addr", round.CreateAddr, "execaddr", a.execaddr, "amount", lockAmount)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	// 将每一轮次的结果保存
	logs = append(logs, &types.ReceiptLog{types.TyLogBlackwhiteLoopInfo, types.Encode(loopResults)})

	return &types.Receipt{types.ExecOk, kv, logs}, nil

}

func (a *action) getWinner(round *types.BlackwhiteRound) ([]*addrResult, *types.ReplyLoopResults) {
	var loopRes types.ReplyLoopResults
	var addresXs []*resultCalc

	loopRes.GameID = round.GetGameID()
	addrRes := round.AddrResult
	loop := int(round.Loop)

	for _, addres := range addrRes {
		if len(addres.ShowSecret) > 0 && len(addres.HashValues) == loop {
			var isBlack []bool
			for _, hash := range addres.HashValues {
				if bytes.Equal(common.Sha256([]byte(addres.ShowSecret+string(black))), hash) {
					isBlack = append(isBlack, true)
				} else {
					isBlack = append(isBlack, false)
				}
			}
			addresX := &resultCalc{
				Addr:    addres.Addr,
				amount:  addres.Amount,
				IsWin:   true,
				IsBlack: isBlack,
			}
			addresXs = append(addresXs, addresX)
		}
	}

	for index := 0; index < loop; index++ {
		blackNum := 0
		whiteNum := 0
		for _, addr := range addresXs {
			if addr.IsWin {
				if addr.IsBlack[index] {
					blackNum++
				} else {
					whiteNum++
				}
			}
		}

		if blackNum < whiteNum {
			for _, addr := range addresXs {
				/*	if addr.IsWin {
					if 0 == blackNum {
						if addr.IsBlack[index] {
							addr.IsWin = false
						}
					} else {
						if false == addr.IsBlack[index] {
							addr.IsWin = false
						}
					}
				}*/
				if addr.IsWin && 0 != blackNum && !addr.IsBlack[index] {
					addr.IsWin = false
				}
			}
		} else if blackNum > whiteNum {
			for _, addr := range addresXs {
				/*	if addr.IsWin {
					if 0 == whiteNum {
						if false == addr.IsBlack[index] {
							addr.IsWin = false
						}
					} else {
						if addr.IsBlack[index] {
							addr.IsWin = false
						}
					}
				}*/
				if addr.IsWin && 0 != whiteNum && addr.IsBlack[index] {
					addr.IsWin = false
				}
			}
		}

		winNum := 0
		var perRes types.PerLoopResult // 每一轮获胜者
		for _, addr := range addresXs {
			if addr.IsWin {
				winNum++
				perRes.Winers = append(perRes.Winers, addr.Addr)
			} else {
				perRes.Losers = append(perRes.Losers, addr.Addr)
			}
		}

		loopRes.Results = append(loopRes.Results, &perRes)

		if 1 == winNum || 2 == winNum {
			break
		}
	}

	var results []*addrResult
	for _, addr := range addresXs {
		if addr.IsWin {
			result := &addrResult{
				addr:   addr.Addr,
				amount: addr.amount,
			}
			results = append(results, result)
		}
	}

	return results, &loopRes
}

func (a *action) getLoser(round *types.BlackwhiteRound) []*addrResult {
	addrRes := round.AddrResult
	wins, _ := a.getWinner(round)

	addMap := make(map[string]bool)
	for _, win := range wins {
		addMap[win.addr] = true
	}

	var results []*addrResult
	for _, addr := range addrRes {
		if ok := addMap[addr.Addr]; !ok {
			result := &addrResult{
				addr:   addr.Addr,
				amount: addr.Amount,
			}
			results = append(results, result)
		}
	}

	return results
}

//状态变化：
// staus == BlackwhiteStatusCreate  (创建，开始游戏）
// status == BlackwhiteStatusPlay (参与)
// status == BlackwhiteStatusShow (展示密钥)
// status == BlackwhiteStatusTimeout (超时退出情况)
// status == BlackwhiteStatusDone (结束情况)

//list 索引保存的方法:
//key=status:addr:gameId
//value=gameId
func (action *action) GetReceiptLog(round *types.BlackwhiteRound) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &types.ReceiptBlackwhiteStatus{}
	if round.Status == gt.BlackwhiteStatusCreate {
		log.Ty = types.TyLogBlackwhiteCreate
		r.PrevStatus = -1
	} else if round.Status == gt.BlackwhiteStatusPlay {
		log.Ty = types.TyLogBlackwhitePlay
		r.PrevStatus = gt.BlackwhiteStatusCreate
	} else if round.Status == gt.BlackwhiteStatusShow {
		log.Ty = types.TyLogBlackwhiteShow
		r.PrevStatus = gt.BlackwhiteStatusPlay
	} else if round.Status == gt.BlackwhiteStatusTimeout {
		log.Ty = types.TyLogBlackwhiteTimeout
		r.PrevStatus = gt.BlackwhiteStatusPlay
	} else if round.Status == gt.BlackwhiteStatusDone {
		log.Ty = types.TyLogBlackwhiteDone
		r.PrevStatus = gt.BlackwhiteStatusShow
	}

	r.GameID = round.GameID
	r.Status = round.Status
	r.Addr = round.CreateAddr

	log.Log = types.Encode(r)
	return log
}

func calcloopNumByPlayer(player int32) int32 {
	a := math.Log2(float64(player))
	a += 0.5
	num := int32(a)
	return num + 1
}
