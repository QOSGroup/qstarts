// Copyright 2018 The QOS Authors

package buyad

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/QOSGroup/qbase/txs"
	qbasetypes "github.com/QOSGroup/qbase/types"
	qostxs "github.com/QOSGroup/qos/txs/transfer"
	"github.com/QOSGroup/qstars/client/utils"
	"github.com/QOSGroup/qstars/config"
	"github.com/QOSGroup/qstars/types"
	"github.com/QOSGroup/qstars/utility"
	"github.com/QOSGroup/qstars/wire"
	"github.com/QOSGroup/qstars/x/common"
	"github.com/QOSGroup/qstars/x/jianqian"
	"log"
	"strconv"
	"strings"
	"time"
)

type ResultBuy struct {
	Code   string          `json:"code"`
	Reason string          `json:"reason,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func InternalError(reason string) ResultBuy {
	return ResultBuy{Code: "-1", Reason: reason}
}

func NewResultBuy(cdc *wire.Codec, code, reason string, res interface{}) ResultBuy {
	var rawMsg json.RawMessage

	if res != nil {
		var js []byte
		js, err := cdc.MarshalJSON(res)
		if err != nil {
			return InternalError(err.Error())
		}
		rawMsg = json.RawMessage(js)
	}

	var result ResultBuy
	result.Result = rawMsg
	result.Code = code
	result.Reason = reason

	return result
}

func (ri ResultBuy) Marshal() string {
	jsonBytes, err := json.MarshalIndent(ri, "", "  ")
	if err != nil {
		log.Printf("BuyAd err:%s", err.Error())
		return InternalError(err.Error()).Marshal()
	}
	return string(jsonBytes)
}

const coinsName = "QOS"
const tempAddr = "address1wmrup5xemdxzx29jalp5c98t7mywulg8wgxxxx"

var shareCommunityAddr = qbasetypes.Address("address1wmrup5xemdxzx29jalp5c98t7mywulg8wgxxxx")

// BuyAdBackground 提交到链上
func BuyAdBackground(cdc *wire.Codec, txb string, timeout time.Duration) string {
	ts := new(txs.TxStd)
	err := cdc.UnmarshalJSON([]byte(txb), ts)
	if err != nil {
		return InternalError(err.Error()).Marshal()
	}

	cliCtx := *config.GetCLIContext().QOSCliContext
	_, commitresult, err := utils.SendTx(cliCtx, cdc, ts)
	if err != nil {
		return InternalError(err.Error()).Marshal()
	}

	height := strconv.FormatInt(commitresult.Height, 10)
	var code, reason string
	var result interface{}

	waittime, err := strconv.Atoi(config.GetCLIContext().Config.WaitingForQosResult)
	if err != nil {
		panic("WaitingForQosResult should be able to convert to integer." + err.Error())
	}
	counter := 0

	for {
		if counter >= waittime {
			log.Println("time out")
			result = "time out"
			break
		}
		resultstr, err := fetchResult(cdc, height, commitresult.Hash.String())
		log.Printf("fetchResult result:%s, err:%+v\n", resultstr, err)
		if err != nil {
			log.Printf("fetchResult error:%s\n", err.Error())
			reason = err.Error()
			break
		}

		if resultstr != "" && resultstr != (BuyTx{}).Name() {
			log.Printf("fetchResult result:[%+v]\n", resultstr)
			rs := []rune(resultstr)
			index1 := strings.Index(resultstr, " ")

			reason = ""
			result = string(rs[index1+1:])
			code = string(rs[:index1])
			break
		}
		time.Sleep(500 * time.Millisecond)
		counter++
	}

	return NewResultBuy(cdc, code, reason, result).Marshal()
}

func fetchResult(cdc *wire.Codec, heigth1 string, tx1 string) (string, error) {
	qstarskey := "heigth:" + heigth1 + ",hash:" + tx1
	d, err := config.GetCLIContext().QSCCliContext.QueryStore([]byte(qstarskey), common.QSCResultMapperName)
	if err != nil {
		return "", err
	}
	if d == nil {
		return "", nil
	}
	var res []byte
	err = cdc.UnmarshalBinaryBare(d, &res)
	if err != nil {
		return "", err
	}
	return string(res), err
}

// BuyAd 投资广告
func BuyAd(cdc *wire.Codec, chainId, articleHash, coins, privatekey string, qosnonce, qscnonce int64) string {
	var result ResultBuy
	result.Code = "0"

	tx, err := buyAd(cdc, chainId, articleHash, coins, privatekey, qosnonce, qscnonce)
	if err != nil {
		log.Printf("buyAd err:%s", err.Error())
		result.Code = "-1"
		result.Reason = err.Error()
		return result.Marshal()
	}

	js, err := cdc.MarshalJSON(tx)
	if err != nil {
		log.Printf("buyAd err:%s", err.Error())
		result.Code = "-1"
		result.Reason = err.Error()
		return result.Marshal()
	}
	result.Result = json.RawMessage(js)

	return result.Marshal()
}

func warpperInvestorTx(articleHash string, amount int64) []qostxs.TransItem {
	var result []qostxs.TransItem

	return result
}

func getCommunityAddr(cdc *wire.Codec) (qbasetypes.Address, error) {
	communityPri := config.GetCLIContext().Config.Community
	if communityPri == "" {
		return nil, errors.New("no community")
	}

	_, addrben32, _ := utility.PubAddrRetrievalFromAmino(communityPri, cdc)
	community, err := types.AccAddressFromBech32(addrben32)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func warpperReceivers(cdc *wire.Codec, articleHash string, amount int64) []qostxs.TransItem {
	var result []qostxs.TransItem
	article := &jianqian.Articles{}

	// 作者地址
	result = append(
		result,
		warpperTransItem(
			article.Authoraddress,
			[]qbasetypes.BaseCoin{{Name: coinsName, Amount: qbasetypes.NewInt(amount * int64(article.ShareAuthor) / 100)}}))

	// 原创作者地址
	result = append(
		result,
		warpperTransItem(
			article.OriginalAuthor,
			[]qbasetypes.BaseCoin{{Name: coinsName, Amount: qbasetypes.NewInt(amount * int64(article.ShareOriginalAuthor) / 100)}}))

	shareCommunityAddr, err := getCommunityAddr(cdc)
	if err == nil {
		// 社区收入比例
		result = append(
			result,
			warpperTransItem(
				shareCommunityAddr,
				[]qbasetypes.BaseCoin{{Name: coinsName, Amount: qbasetypes.NewInt(amount * int64(article.ShareCommunity) / 100)}}))

		// 投资者收入分配
		shareInvestorTotal := amount * int64(article.ShareCommunity) / 100
		result = append(result, warpperInvestorTx(articleHash, shareInvestorTotal)...)
	}

	return result
}

// buyAd 投资广告
func buyAd(cdc *wire.Codec, chainId, articleHash, coins, privatekey string, qosnonce, qscnonce int64) (*txs.TxStd, error) {
	cs, err := types.ParseCoins(coins)
	if err != nil {
		return nil, err
	}

	if len(cs) != 1 {
		return nil, errors.New("one coin need")
	}

	for _, v := range cs {
		if v.Denom != coinsName {
			return nil, fmt.Errorf("only support %s", coinsName)
		}
	}

	var amount int64
	_, addrben32, priv := utility.PubAddrRetrievalFromAmino(privatekey, cdc)
	buyer, err := types.AccAddressFromBech32(addrben32)
	var ccs []qbasetypes.BaseCoin
	for _, coin := range cs {
		amount = coin.Amount.Int64()
		ccs = append(ccs, qbasetypes.BaseCoin{
			Name:   coin.Denom,
			Amount: qbasetypes.NewInt(coin.Amount.Int64()),
		})
	}
	qosnonce += 1
	var transferTx qostxs.TxTransfer
	transferTx.Senders = []qostxs.TransItem{warpperTransItem(buyer, ccs)}
	transferTx.Receivers = warpperReceivers(cdc, articleHash, amount)
	gas := qbasetypes.NewInt(int64(0))
	stx := txs.NewTxStd(transferTx, config.GetCLIContext().Config.QOSChainID, gas)
	signature, _ := stx.SignTx(priv, qosnonce, config.GetCLIContext().Config.QSCChainID)
	stx.Signature = []txs.Signature{txs.Signature{
		Pubkey:    priv.PubKey(),
		Signature: signature,
		Nonce:     qosnonce,
	}}

	qscnonce += 1
	it := &BuyTx{}
	it.ArticleHash = []byte(articleHash)
	it.Std = stx
	tx2 := txs.NewTxStd(it, config.GetCLIContext().Config.QSCChainID, stx.MaxGas)
	signature2, _ := tx2.SignTx(priv, qscnonce, config.GetCLIContext().Config.QSCChainID)
	tx2.Signature = []txs.Signature{txs.Signature{
		Pubkey:    priv.PubKey(),
		Signature: signature2,
		Nonce:     qscnonce,
	}}

	return tx2, nil
}

func warpperTransItem(addr qbasetypes.Address, coins []qbasetypes.BaseCoin) qostxs.TransItem {
	var ti qostxs.TransItem
	ti.Address = addr
	ti.QOS = qbasetypes.NewInt(0)

	for _, coin := range coins {
		if coin.Name == "qos" {
			ti.QOS = ti.QOS.Add(coin.Amount)
		} else {
			ti.QSCs = append(ti.QSCs, &coin)
		}
	}

	return ti
}
