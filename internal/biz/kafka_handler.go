package biz

import (
	"errors"
	"fmt"
	"log/slog"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"

	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
)

func (uc *WalletUC) handleSenderConfirm(kmsg kafka.Message) error {
	req := domain.TransactionReq{}
	if err := msgpack.Unmarshal(kmsg.Value, &req); err != nil {
		return err
	}
	slog.Debug("sender confirm received kafka msg", "req", req)
	if err := uc.repo.Withdraw(req); err != nil {
		if errors.Is(err, errs.ErrDuplicate) { // same idempotency key
			return nil
		}
		if errors.Is(err, errs.ErrInsufficientBalance) {
			return uc.tccCancel(req)
		}
		return err
	}
	// receiver confirm
	return uc.tccConfirm(req)
}

func (uc *WalletUC) handleReceiverConfirm(kmsg kafka.Message) error {
	req := domain.TransactionReq{}
	if err := msgpack.Unmarshal(kmsg.Value, &req); err != nil {
		return err
	}
	// convert the sender's transfer request to a receiver deposit request
	if req.TargetID == nil {
		return fmt.Errorf("transaction req %s no target id for receiver ,%w", req.IdempotencyKey, errs.ErrInvalidParam)
	}
	recvReq := domain.TransactionReq{
		IdempotencyKey: req.IdempotencyKey,
		UserID:         *req.TargetID,
		TargetID:       &req.UserID,
		Amt:            req.Amt,
		Type:           domain.Deposit,
	}
	slog.Debug("receiver confirm received kafka msg", "req", req, "recvReq", recvReq)
	return uc.repo.Deposit(recvReq)
}

func (uc *WalletUC) handleCancel(kmsg kafka.Message) error {
	req := domain.TransactionReq{}
	if err := msgpack.Unmarshal(kmsg.Value, &req); err != nil {
		return err
	}
	slog.Debug("cancel received kafka msg", "req", req)
	return uc.repo.CancelFrozenBalance(req)
}
