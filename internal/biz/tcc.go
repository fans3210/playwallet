package biz

import (
	"errors"
	"fmt"
	"log/slog"

	"playwallet/internal/cfgs"
	"playwallet/internal/domain"
	"playwallet/pkg/errs"
)

// 1. try: create frozen balance record, pub kafka msg to `sender_confirm` topic
// 2. confirm: once received kafka msg, if have enough balance, create transaction record, mark `confirmed`, pub kafka msg to `receiver confirm` topic,
// otherwise, pub kafka msg to `cancel` topic
// 3. cancel: once received kafka msg, if not have enough balance, mark the frozen_balance record `cancelled`

func (uc *WalletUC) tccTry(req domain.TransactionReq) error {
	shouldContinue := true

	defer func() {
		if !shouldContinue {
			return
		}
		sendConfirm, ok := uc.senders[cfgs.TpcKeySenderConfirm]
		if !ok {
			slog.Error("no sender for topic for confirm")
			return
		}
		kMsg, err := req.ToKafkaMsg()
		if err != nil {
			slog.Error("fail to convert TransactionReq to kafka msg", "req", req, "err", err)
			return
		}
		if err := sendConfirm.SendMsg(*kMsg); err != nil {
			slog.Error("fail to send confirm kafka msg", "err", err)
		}
	}()

	if err := uc.repo.CreateFrozenBalance(req); err != nil {
		if !errors.Is(err, errs.ErrDuplicate) {
			shouldContinue = false
			return err
		}
		slog.Error("failed to create frozen balance for transfer due to duplicate, continue the `try` step \n", "err", err)
		return err
	}
	return nil
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
