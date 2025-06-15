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

func (uc *WalletUC) handleSenderConfirm(kmsg kafka.Message) (err error) {
	req := domain.TransactionReq{}
	if err := msgpack.Unmarshal(kmsg.Value, &req); err != nil {
		return fmt.Errorf("failed to Unmarshal kafka msg: %w", err)
	}
	slog.Debug("sender confirm received kafka msg", "req", req)
	// WARN: disable retry, would impact performance due to kafka msg, simply cancel, as long as no over deduction
	// withDrawErr := tools.Retry(3, func() error {
	// 	return uc.repo.Withdraw(req)
	// })
	withDrawErr := uc.repo.Withdraw(req)
	if err := withDrawErr; err != nil {
		if errors.Is(err, errs.ErrDuplicate) { // same idempotency key
			// continue receiver confirm for IdempotencyKey issue
			return uc.tccConfirm(req)
		}
		if errors.Is(err, errs.ErrInsufficientBalance) {
			slog.Warn("handle sender confirm msg ErrInsufficientBalance, prepare to cancel, req", "req", req, "insufficnentbalanceerr", err)
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
		return fmt.Errorf("failed to Unmarshal kafka msg: %w", err)
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
	if err := uc.repo.Deposit(recvReq); err != nil {
		if errors.Is(err, errs.ErrDuplicate) { // same idempotency key
			return nil
		}
		return err
	}
	return nil
}

func (uc *WalletUC) handleCancel(kmsg kafka.Message) error {
	req := domain.TransactionReq{}
	if err := msgpack.Unmarshal(kmsg.Value, &req); err != nil {
		return fmt.Errorf("failed to Unmarshal kafka msg: %w", err)
	}
	slog.Debug("cancel received kafka msg", "req", req)
	if err := uc.repo.CancelFrozenBalance(req); err != nil {
		slog.Error("cancel failed", "req", req)
		return err
	}
	slog.Error("cancel succeed", "req", req)
	return nil
}
