// Copyright 2018 The QOS Authors

package jianqian

import (
	"github.com/QOSGroup/qbase/mapper"
	qbasetypes "github.com/QOSGroup/qbase/types"
	"log"
	"time"
)

const (
	InvestMapperName = "investad"
)

func GetInvestKey(article []byte, user qbasetypes.Address) []byte {
	return append(article, user...)
}

// Investor 投资者
type Investor struct {
	Address    qbasetypes.Address `json:"address"`    // 投资者地址
	Invest     qbasetypes.BigInt  `json:"investad"`   // 投资金额
	Revenue    qbasetypes.BigInt  `json:"revenue"`    // 投资收益
	InvestTime time.Time          `json:"investTime"` // 投资时间
}

type InvestMapper struct {
	*mapper.BaseMapper
}

var _ mapper.IMapper = (*InvestMapper)(nil)

func (im *InvestMapper) Copy() mapper.IMapper {
	cpyMapper := &InvestMapper{}
	cpyMapper.BaseMapper = im.BaseMapper.Copy()
	return cpyMapper
}

func NewInvestMapper(mapperName string) *InvestMapper {
	var investMapper = InvestMapper{}
	investMapper.BaseMapper = mapper.NewBaseMapper(nil, mapperName)
	return &investMapper
}

func (im *InvestMapper) SaveKV(key string, value string) {
	im.BaseMapper.Set([]byte(key), value)
}

func (im *InvestMapper) GetKey(key string) (v string) {
	im.BaseMapper.Get([]byte(key), &v)
	return
}

// Get 查询用户投资情况
func (im *InvestMapper) GetInvestor(key []byte) (Investor, bool) {
	log.Printf("jianqian.GetInvestor key:%+v", key)
	var result Investor
	ok := im.Get(key, &result)
	return result, ok
}

// Set 添加用户投资
func (im *InvestMapper) SetInvestor(key []byte, i Investor) {
	log.Printf("jianqian.SetInvestor key:%+v, v:%+v", key, i)
	im.Set(key, i)
	return
}

// Iterator
func (im *InvestMapper) AllInvestors(article []byte) []Investor {
	log.Printf("jianqian.AllInvestors article:%+v", article)

	var investors []Investor

	im.Iterator(article, func(val []byte) (stop bool) {
		log.Printf("jianqian.AllInvestors Iterator")
		var investor Investor
		im.DecodeObject(val, &investor)
		log.Printf("jianqian.AllInvestors Iterator val:%+v, investor:%+v", val, investor)
		investors = append(investors, investor)
		return false
	})

	return investors
}
