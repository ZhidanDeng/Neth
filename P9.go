package main

import (
	"github.com/ethereum/collector"
	"math/big"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RegisterInfo struct {
	PluginName string            `json:"pluginname"`
	OpCode     map[string]string `json:"option"`
}

//检测结果导出的格式
type DaoInfo struct {
	BlockNumber string `json:"blocknumber"`
	TxHash      string `json:"txhash"`
	IntegerType string `json:"integertype"`
	Status      string `json:"status"`
}

//全局变量
var (
	txhash      string
	blocknumber string
	integerType string
	status      string
	max         *big.Int
)

func Register() []byte {
	var data = RegisterInfo{
		PluginName: "P9",
		OpCode:     map[string]string{"ADD": "Handle_ADD", "SUB": "Handle_SUB", "MUL": "Handle_MUL"},
	}

	retInfo, err := json.Marshal(&data)
	if err != nil {
		return []byte{}
	}
	return retInfo
}

//加法溢出
func Handle_ADD(m *collector.AllCollector) (byte, string) {
	bx, by, bres := getArgsAndRes(m)
	txhash = m.TransInfo.TxHash
	blocknumber = m.TransInfo.BlockNumber
	integerType = "ADD"
	if bx == nil || by == nil || bres == nil {
		status = "Failed to detect"
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
	}
	status = ""
	//如果res < x || res < y说明溢出
	if bres.Cmp(bx) == -1 || bres.Cmp(by) == -1 {
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
		return 0x00, ""
	}
	return 0x00, ""
}

//乘法溢出
func Handle_MUL(m *collector.AllCollector) (byte, string) {
	bx, by, bres := getArgsAndRes(m)
	txhash = m.TransInfo.TxHash
	blocknumber = m.TransInfo.BlockNumber
	integerType = "MUL"
	if bx == nil || by == nil || bres == nil {
		status = "Failed to detect"
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
	}
	status = ""
	if bres.Cmp(big.NewInt(0)) == 0 {
		return 0x00, ""
	}
	t := new(big.Int)
	//  bres/bx supposed to be by
	t.Div(bres, bx)
	if t.Cmp(by) != 0 {
		//不正常
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
	}
	return 0x00, ""
}

//减法溢出
func Handle_SUB(m *collector.AllCollector) (byte, string) {
	bx, by, bres := getArgsAndRes(m)
	txhash = m.TransInfo.TxHash
	blocknumber = m.TransInfo.BlockNumber
	integerType = "SUB"
	if bx == nil || by == nil || bres == nil {
		status = "Failed to detect"
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
	}
	status = ""
	if bres.Cmp(big.NewInt(0)) == 0 {
		return 0x00, ""
	}
	if bx.Cmp(by) < 0 {
		//bx < by
		max = by
	} else if bx.Cmp(by) > 0 {
		max = bx
	}
	if max.Cmp(bres) < 0 {
		//不正常
		result := processInteger()
		if result != "" {
			return 0x01, result
		}
	}
	return 0x00, ""
}

func getArgsAndRes(m *collector.AllCollector) (*big.Int, *big.Int, *big.Int) {
	inOutInfo := m.InsInfo.OpInOut
	args := inOutInfo.OpArgs
	if len(args) == 2 {
		var x = args[0]
		var y = args[1]
		var res = inOutInfo.OpResult //加减除都是y，乘法是x
		bx := new(big.Int)
		by := new(big.Int)
		bres := new(big.Int)
		//string => big.Int
		bx, _ = bx.SetString(x, 10)
		by, _ = by.SetString(y, 10)
		bres, _ = bres.SetString(res, 10)
		return bx, by, bres
	}
	return nil, nil, nil

}

func processInteger() string {
	daoInfo := DaoInfo{
		blocknumber,
		txhash,
		integerType,
		status,
	}

	daoJsonData, err := json.Marshal(daoInfo)
	if err != nil {
		panic(err)
	}

	return string(daoJsonData)
}
