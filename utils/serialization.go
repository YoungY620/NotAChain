package utils

import (
	"encoding/json"
	"neochain/common"
)

func JsonToTxDefMsg(d string) (*common.TxDefMsg, error) {
	var msg common.TxDefMsg
	err := json.Unmarshal([]byte(d), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func TxDefMsgToTransaction(msg *common.TxDefMsg) *common.Transaction {
	return common.NewTransaction(msg.IdxFrom, msg.IdxTo)
}

func TxDefMsgToJson(msg *common.TxDefMsg) (string, error) {
	str, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(str), nil
}
