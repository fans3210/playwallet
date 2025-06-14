package biz

import (
	"fmt"

	"playwallet/internal/cfgs"
	"playwallet/internal/domain"
)

// 1. try: create frozen balance record, pub kafka msg to `sender_confirm` topic
// 2. confirm: once received kafka msg, if have enough balance, create transaction record, mark `confirmed`, pub kafka msg to `receiver confirm` topic,
// otherwise, pub kafka msg to `cancel` topic
// 3. cancel: once received kafka msg, if not have enough balance, mark the frozen_balance record `cancelled`

func (uc *WalletUC) tccTry(req domain.TransactionReq) error {
	if err := uc.repo.CreateFrozenBalance(req); err != nil {
		return err
	}
	sendConfirm, ok := uc.senders[cfgs.TpcKeySenderConfirm]
	if !ok {
		return fmt.Errorf("no sender for topic for confirm")
	}
	kMsg, err := req.ToKafkaMsg()
	if err != nil {
		return err
	}
	return sendConfirm.SendMsg(*kMsg)
}

func (uc *WalletUC) tccCancel(req domain.TransactionReq) error {
	sendCancel, ok := uc.senders[cfgs.TpcKeyCancel]
	if !ok {
		return fmt.Errorf("no sender found for topic for cancel")
	}
	kMsg, err := req.ToKafkaMsg()
	if err != nil {
		return err
	}
	return sendCancel.SendMsg(*kMsg)
}

// receiver confirm
func (uc *WalletUC) tccConfirm(req domain.TransactionReq) error {
	sendRecvConfirm, ok := uc.senders[cfgs.TpcKeyReceiverConfirm]
	if !ok {
		return fmt.Errorf("no sender found for topic for receiver confirm")
	}
	kMsg, err := req.ToKafkaMsg()
	if err != nil {
		return err
	}
	return sendRecvConfirm.SendMsg(*kMsg)
}
