package tests

import (
	"errors"
	"net/http"
	"testing"

	"playwallet/internal/domain"
)

func TestCredit(t *testing.T) {
	endpoint, _, teardown := provisionTestApp(t)
	defer teardown(t)
	uid := 1
	idpKey := "transfer1"
	firstDepositAmt := 10
	secondDepositAmt := 100
	req := domain.DepositReq{
		UserID:         int64(uid),
		Amt:            int64(firstDepositAmt),
		IdempotencyKey: idpKey,
	}
	status, err := makeDepositReq(endpoint, req)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("unexpected status code, expect 200, actual: %d\n", status)
	}

	// expect error for same idempotency key for second time deposit if is same user
	// userid + idempotency key should be unique for transactions,
	status, err = makeDepositReq(endpoint, req)
	if err == nil {
		t.Fatalf("expect error for same idempotency key for same user: %s\n", idpKey)
	}
	if !errors.Is(err, errNon200Status) {
		t.Fatalf("incorrect error type, expect: %s, actual: %s\n", errNon200Status, err)
	}
	if status != http.StatusConflict {
		t.Fatalf("unexpected status code for duplicate idempotency key case , expect: %d, actual: %d\n", http.StatusConflict, status)
	}

	// if user2 use same idempotency key, can deposit without issue since idempotency+userid is unique, and won't affect user1's balance
	reqUser2 := req
	reqUser2.UserID = 2
	status, err = makeDepositReq(endpoint, reqUser2)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("unexpected status code, expect 200, actual: %d\n", status)
	}

	// topup again and check balance
	idpKey = "transfer2"
	req2 := domain.DepositReq{
		UserID:         1,
		Amt:            int64(secondDepositAmt),
		IdempotencyKey: idpKey,
	}
	status, err = makeDepositReq(endpoint, req2)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("unexpected status code, expect 200, actual: %d\n", status)
	}

	status, balanceInfo, err := makeCheckBalanceReq(endpoint, 1)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("unexpected status code for balance check after deposit, expect 200, actual: %d\n", status)
	}
	if !balanceInfo.IsValid() || balanceInfo.UserID != 1 {
		t.Fatalf("invalid balance info: %+v\n", balanceInfo)
	}
	if balanceInfo.AvailableBalance != int64(firstDepositAmt+secondDepositAmt) {
		t.Fatalf("incorrect available balance amt after deposit, expect: %d, actual: %d\n", firstDepositAmt+secondDepositAmt, balanceInfo.AvailableBalance)
	}
}

// TODO: test idempotency key

// TODO:  for debit, test idempotency key, test concurrent debit to simulate mallisious user who expect to get more cash

// TODO: for transfer, test idempotency key in each step, the try, confirm, cancel step,
// TODO: for transfer, test concurrent transfer to simulate mallisious user who expect to get more amt from source user
// TODO: for transfer, test concurrent transfer+deposit+debit, check the correctness of the result balance
