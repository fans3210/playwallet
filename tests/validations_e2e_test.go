package tests

import (
	"errors"
	"net/http"
	"testing"

	"playwallet/internal/domain"
)

func TestCheckBalanceValidation(t *testing.T) {
	endpoint, _, teardown := provisionTestApp(t)
	defer teardown(t)
	nonExistUserID := 10101010
	status, _, err := makeCheckBalanceReq(endpoint, int64(nonExistUserID))
	if err == nil {
		t.Fatalf("expect error for non exist userid: %d\n", nonExistUserID)
	}
	if !errors.Is(err, errNon200Status) {
		t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("unexpected response status: %d, expect 404 not found\n", status)
	}
}

// TODO: add withdraw test, add transfer test for more validation cases
func TestMakeTransactionValidation(t *testing.T) {
	cases := []struct {
		desc          string
		req           domain.TransactionReq
		expectErrCode int
	}{
		{
			desc: "make transaction missing transaction type",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         1,
				Amt:            10,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "deposit non exist user",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         100100100,
				Amt:            10,
				Type:           domain.Deposit,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "withdraw non exist user",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         100100100,
				Amt:            10,
				Type:           domain.Withdraw,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "deposit negative amt",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         1,
				Amt:            -10,
				Type:           domain.Deposit,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "withdraw negative amt",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         1,
				Amt:            -10,
				Type:           domain.Withdraw,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "deposit empty idempotency key",
			req: domain.TransactionReq{
				UserID: 1,
				Amt:    10,
				Type:   domain.Deposit,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "withdraw empty idempotency key",
			req: domain.TransactionReq{
				UserID: 1,
				Amt:    10,
				Type:   domain.Withdraw,
			},
			expectErrCode: http.StatusBadRequest,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			endpoint, _, teardown := provisionTestApp(t)
			defer teardown(t)
			status, err := makeTransactionReq(endpoint, c.req)
			if err == nil {
				t.Fatalf("invalid request should have error")
			}
			if !errors.Is(err, errNon200Status) {
				t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
			}
			if status != c.expectErrCode {
				t.Fatalf("unexpected response status: %d, expect %d\n", status, c.expectErrCode)
			}
		})
	}
}
