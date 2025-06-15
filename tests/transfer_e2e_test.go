package tests

import (
	"context"
	"testing"
	"time"

	"playwallet/internal/domain"
)

// after try, send another kafka msg with same content, should forbid creating another record due to idemptency key
func TestTransferTry(t *testing.T) {
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	uid1, uid2 := int64(1), int64(2)
	startAmt := uint64(100)
	transferAmt1, transferAmt2 := int64(10), int64(50)
	idpK1, idpK2 := "hi", "hi2"

	addTestData(t, db,
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
	)
	if _, err := makeTransactionReq(endpoint, domain.TransactionReq{
		IdempotencyKey: idpK1,
		UserID:         uid1,
		TargetID:       &uid2,
		Amt:            transferAmt1,
		Type:           domain.Transfer,
	}); err != nil {
		t.Fatalf("fail to make transfer req 1: %s\n", err)
	}

	// send antoher trasnfer with same idemptency key, expect error, and should cancel and thus won't affect the result balance
	if _, err := makeTransactionReq(endpoint, domain.TransactionReq{
		IdempotencyKey: idpK1,
		UserID:         uid1,
		TargetID:       &uid2,
		Amt:            transferAmt1,
		Type:           domain.Transfer,
	}); err == nil {
		t.Fatalf("expect error if making another transaction with same idemptency key: %s\n", idpK1)
	}

	if _, err := makeTransactionReq(endpoint, domain.TransactionReq{
		IdempotencyKey: idpK2,
		UserID:         uid2,
		TargetID:       &uid1,
		Amt:            transferAmt2,
		Type:           domain.Transfer,
	}); err != nil {
		t.Fatalf("fail to make transfer req 2: %s\n", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("time out, transfer not in desired state")
		case <-ticker.C:
			_, balaceInfo, err := makeCheckBalanceReq(endpoint, uid1)
			if err != nil {
				t.Fatal(err)
			}
			expected := int64(startAmt) - transferAmt1 + transferAmt2
			if balaceInfo.AvailableBalance != expected {
				t.Logf("incorrect balance for user1 after transfer, expect: %d, actual: %d\n", expected, balaceInfo.AvailableBalance)
				// won't return and wait for next round check since is async and eventually consistent
				continue
			}
			_, balaceInfo, err = makeCheckBalanceReq(endpoint, uid2)
			if err != nil {
				t.Fatal(err)
			}
			expected = int64(startAmt) + transferAmt1 - transferAmt2
			if balaceInfo.AvailableBalance != expected {
				t.Logf("incorrect balance for user2 after transfer, expect: %d, actual: %d\n", expected, balaceInfo.AvailableBalance)
				// won't return and wait for next round check since is async and eventually consistent
				continue
			}
			return
		}
	}
}
