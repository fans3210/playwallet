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

func TestGetTransactionsValidation(t *testing.T) {
	endpoint, _, teardown := provisionTestApp(t)
	defer teardown(t)
	nonExistUserID := 10101010
	status, _, err := makeGetTransactionsReq(endpoint, int64(nonExistUserID), nil)
	if err == nil {
		t.Fatalf("expect error for non exist userid: %d\n", nonExistUserID)
	}
	if !errors.Is(err, errNon200Status) {
		t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
	}
	if status != http.StatusNotFound {
		t.Fatalf("unexpected response status: %d, expect 404 not found\n", status)
	}

	invalidPageOpt := &domain.PageOpt{Page: -1, PerPage: -1}
	status, _, err = makeGetTransactionsReq(endpoint, 1, invalidPageOpt)
	if err == nil {
		t.Fatalf("expect error for invalid page opt: %+v\n", invalidPageOpt)
	}
	if !errors.Is(err, errNon200Status) {
		t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
	}
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected response status: %d, expect bad request for invalid page opt\n", status)
	}
}

func TestMakeTransactionValidation(t *testing.T) {
	uid1, uid2, uidNonExist := int64(1), int64(2), int64(10010101010)

	cases := []struct {
		desc          string
		req           domain.TransactionReq
		expectErrCode int
	}{
		{
			desc: "make transaction missing transaction type",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uid1,
				Amt:            10,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "deposit non exist user",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uidNonExist,
				Amt:            10,
				Type:           domain.Deposit,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "withdraw non exist user",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uidNonExist,
				Amt:            10,
				Type:           domain.Withdraw,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "transfer non exist user 1",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uidNonExist,
				TargetID:       &uid2,
				Amt:            10,
				Type:           domain.Transfer,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "transfer non exist user 2",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uid1,
				TargetID:       &uidNonExist,
				Amt:            10,
				Type:           domain.Transfer,
			},
			expectErrCode: http.StatusNotFound,
		},
		{
			desc: "deposit negative amt",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uid1,
				Amt:            -10,
				Type:           domain.Deposit,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "withdraw negative amt",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uid1,
				Amt:            -10,
				Type:           domain.Withdraw,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "transfer negative amt",
			req: domain.TransactionReq{
				IdempotencyKey: "testkey",
				UserID:         uid1,
				Amt:            -10,
				Type:           domain.Transfer,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "deposit empty idempotency key",
			req: domain.TransactionReq{
				UserID: uid1,
				Amt:    10,
				Type:   domain.Deposit,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "withdraw empty idempotency key",
			req: domain.TransactionReq{
				UserID: uid1,
				Amt:    10,
				Type:   domain.Withdraw,
			},
			expectErrCode: http.StatusBadRequest,
		},
		{
			desc: "transfer empty idempotency key",
			req: domain.TransactionReq{
				UserID: uid1,
				Amt:    10,
				Type:   domain.Transfer,
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
