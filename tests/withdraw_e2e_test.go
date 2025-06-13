package tests

import (
	"errors"
	"net/http"
	"testing"

	"playwallet/internal/domain"
)

func TestWithdraw(t *testing.T) {
	t.SkipNow()
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	// prepare data
	uid1, uid2 := int64(1), int64(2)
	startAmt := uint64(100)
	idpKey1, idpKey2 := "transfer1", "transfer2"
	amt1, amt2 := int64(10), int64(20)

	if err := addTestData(t, db,
		action{
			userID:  uid1,
			actType: deposit,
			amt:     startAmt,
		},
		action{
			userID:  uid2,
			actType: deposit,
			amt:     startAmt,
		},
	); err != nil {
		t.Fatal(err)
	}

	// withdraw amt that exceed the available amt
	req := domain.TransactionReq{
		UserID:         uid1,
		Amt:            int64(startAmt) + 10000,
		IdempotencyKey: idpKey1,
		Type:           domain.Withdraw,
	}
	status, err := makeTransactionReq(endpoint, req)
	if err == nil {
		t.Fatal("expect error if withdraw amt > available balance")
	}
	if status != http.StatusForbidden {
		t.Fatalf("unexpected status code, expect 403 forbidden, actual: %d\n", status)
	}
	_, balanceInfo, err := makeCheckBalanceReq(endpoint, uid1)
	if err != nil {
		t.Fatal(err)
	}
	if balanceInfo.AvailableBalance != int64(startAmt) {
		t.Fatalf("incorrect available balance amt, expect: %d, actual: %d\n", startAmt, balanceInfo.AvailableBalance)
	}

	// make first widhraws
	req = domain.TransactionReq{
		UserID:         uid1,
		Amt:            amt1,
		IdempotencyKey: idpKey1,
		Type:           domain.Withdraw,
	}
	if _, err := makeTransactionReq(endpoint, req); err != nil {
		t.Fatal(err)
	}
	_, balanceInfo, err = makeCheckBalanceReq(endpoint, uid1)
	if err != nil {
		t.Fatal(err)
	}
	if balanceInfo.AvailableBalance != int64(startAmt)-amt1 {
		t.Fatalf("incorrect available balance amt after withdraw, expect: %d, actual: %d\n", int64(startAmt)-amt1, balanceInfo.AvailableBalance)
	}

	// expect error for same idempotency key for second time withdraw if is same user
	// userid + idempotency key should be unique for transactions,
	status, err = makeTransactionReq(endpoint, req)
	if err == nil {
		t.Fatalf("expect error for same idempotency key for same user: %s\n", idpKey1)
	}
	if !errors.Is(err, errNon200Status) {
		t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
	}
	if status != http.StatusConflict {
		t.Fatalf("unexpected status code for duplicate idempotency key case , expect: %d, actual: %d\n", http.StatusConflict, status)
	}

	// if user2 use same idempotency key, can withdraw without issue since idempotency+userid is unique, and won't affect user1's balance
	reqUser2 := req
	reqUser2.UserID = uid2
	if _, err := makeTransactionReq(endpoint, reqUser2); err != nil {
		t.Fatal(err)
	}

	// withdraw again and check balance
	req2 := domain.TransactionReq{
		UserID:         uid1,
		Amt:            int64(amt2),
		IdempotencyKey: idpKey2,
		Type:           domain.Withdraw,
	}
	if _, err = makeTransactionReq(endpoint, req2); err != nil {
		t.Fatal(err)
	}
	_, balanceInfo, err = makeCheckBalanceReq(endpoint, uid1)
	if err != nil {
		t.Fatal(err)
	}
	if balanceInfo.AvailableBalance != int64(startAmt)-amt1-amt2 {
		t.Fatalf("incorrect available balance amt after withdraw, expect: %d, actual: %d\n", int64(startAmt)-amt1-amt2, balanceInfo.AvailableBalance)
	}
}
