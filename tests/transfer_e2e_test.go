package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"playwallet/internal/domain"
)

func TestTransferNotEnoughBalanceShouldCancel(t *testing.T) {
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	uid1, uid2 := int64(1), int64(2)
	startAmt := uint64(25)
	transferAmt := int64(1000)
	idpK := "hi"
	addTestData(t, db,
		action{
			userID:  uid1,
			actType: deposit,
			amt:     startAmt,
		},
	)
	if _, err := makeTransactionReq(endpoint, domain.TransactionReq{
		IdempotencyKey: idpK,
		UserID:         uid1,
		TargetID:       &uid2,
		Amt:            transferAmt,
		Type:           domain.Transfer,
	}); err != nil {
		t.Fatalf("fail to make transfer req 1: %s\n", err)
	}

	// after trasaction cancel, user's balance should not change
	assertBalance(t, endpoint, uid1, uid2, int64(startAmt), 0, 30)
}

// after try, send another kafka msg with same content, should forbid creating another record due to idemptency key
func TestTransferHappyPath(t *testing.T) {
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

	// send antoher transfer with same idemptency key, expect error, and should cancel and thus won't affect the result balance
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

	expected1 := int64(startAmt) - transferAmt1 + transferAmt2
	expected2 := int64(startAmt) + transferAmt1 - transferAmt2
	assertBalance(t, endpoint, uid1, uid2, expected1, expected2, 10)
}

func TestConcurrentTransfer(t *testing.T) {
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	uid1, uid2 := int64(1), int64(2)
	startAmt := uint64(100)
	transferAmt := int64(15)
	addTestData(t, db,
		action{
			userID:  uid1,
			actType: deposit,
			amt:     startAmt,
		},
	)
	N := 100
	wg := sync.WaitGroup{}
	sema := make(chan struct{}, 50) // WARN: by default postgres allow max 100 client conns at the same time, need to limit to be less than 100
	for i := 1; i <= N; i++ {
		idpK := fmt.Sprintf("hihi:%d", i)
		wg.Add(1)
		go func() {
			sema <- struct{}{}
			defer func() {
				wg.Done()
				<-sema
			}()
			_, _ = makeTransactionReq(endpoint, domain.TransactionReq{
				IdempotencyKey: idpK,
				UserID:         uid1,
				TargetID:       &uid2,
				Amt:            transferAmt,
				Type:           domain.Transfer,
			})
		}()
	}
	wg.Wait()
	// WARN:
	// ideally, user 1 should have 10 left, but sometimes due to race condition -
	// when there is only 25 left, sometimes two concurrent transactions would both fail to deduct due to conflict, so there would be two cancelled transactions
	// and 25 would left
	expected1 := max(int64(startAmt)-int64(N-1)*transferAmt, int64(startAmt)%transferAmt+transferAmt)
	expected1Opt := max(int64(startAmt)-int64(N)*transferAmt, int64(startAmt)%transferAmt+transferAmt)
	expected2 := min(0+transferAmt*int64(N-1), int64(startAmt)-(int64(startAmt)%transferAmt+transferAmt))
	expected2Opt := min(0+transferAmt*int64(N), int64(startAmt)-(int64(startAmt)%transferAmt+transferAmt))
	// WARN: must make sure to have enough timeout to make all the kafka msgs consumed successfully
	assertBalance(t, endpoint, uid1, uid2, expected1, expected2, 30, expected1Opt, expected2Opt)
}

// assume user have 2 accounts, attempt to DDOS and make the same transfer and expect receiver account to get more than what he sent
func TestMalliciousTransfer(t *testing.T) {
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	uid1, uid2 := int64(1), int64(2)
	startAmt := uint64(100)
	transferAmt := int64(10)
	idpK1 := "hi"
	addTestData(t, db,
		action{
			userID:  uid1,
			actType: deposit,
			amt:     startAmt,
		},
	)
	wg := sync.WaitGroup{}
	sema := make(chan struct{}, 50) // WARN: by default postgres allow max 100 client conns at the same time, need to limit to be less than 100
	for i := 1; i <= 500; i++ {
		wg.Add(1)
		go func() {
			sema <- struct{}{}
			defer func() {
				wg.Done()
				<-sema
			}()
			_, _ = makeTransactionReq(endpoint, domain.TransactionReq{
				IdempotencyKey: idpK1,
				UserID:         uid1,
				TargetID:       &uid2,
				Amt:            transferAmt,
				Type:           domain.Transfer,
			})
		}()
	}
	wg.Wait()
	assertBalance(t, endpoint, uid1, uid2, int64(startAmt)-transferAmt, transferAmt, 60)
}

func assertBalance(t *testing.T, endpoint string, uid1, uid2, expect1, expect2 int64, timeoutSecs int, optVals ...int64) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("time out, transfer balance not in desired state")
		case <-ticker.C:
			_, balaceInfo, err := makeCheckBalanceReq(endpoint, uid1)
			if err != nil {
				t.Fatal(err)
			}
			if !balaceInfo.IsValid() {
				t.Fatalf("invalid balance info: %+v\n", balaceInfo)
			}
			if balaceInfo.AvailableBalance != expect1 {
				if len(optVals) >= 2 {
					if balaceInfo.AvailableBalance != optVals[0] {
						t.Logf("incorrect balance for user1 after transfer, expect: %d or %d, actual: %d, info: %+v\n", expect1, optVals[0], balaceInfo.AvailableBalance, balaceInfo)
						continue
					}
				} else {
					t.Logf("incorrect balance for user1 after transfer, expect: %d, actual: %d, info: %+v\n", expect1, balaceInfo.AvailableBalance, balaceInfo)
					// won't return and wait for next round check since is async and eventually consistent
					continue
				}
			}
			t.Logf("balance check for user 1 passed, expact1: %d, actual1: %d\n", expect1, balaceInfo.AvailableBalance)
			_, balaceInfo, err = makeCheckBalanceReq(endpoint, uid2)
			if err != nil {
				t.Fatal(err)
			}
			if !balaceInfo.IsValid() {
				t.Fatalf("invalid balance info: %+v\n", balaceInfo)
			}
			if balaceInfo.AvailableBalance != expect2 {
				if len(optVals) >= 2 {
					if balaceInfo.AvailableBalance != optVals[1] {
						t.Logf("incorrect balance for user2 after transfer, expect: %d or %d, actual: %d, info: %+v\n", expect2, optVals[1], balaceInfo.AvailableBalance, balaceInfo)
						continue
					}
				} else {
					t.Logf("incorrect balance for user2 after transfer, expect: %d, actual: %d, info: %+v\n", expect2, balaceInfo.AvailableBalance, balaceInfo)
					// won't return and wait for next round check since is async and eventually consistent
					continue
				}
			}
			t.Logf("balance check for user 2 passed, expact2: %d, actual2: %d\n", expect2, balaceInfo.AvailableBalance)
			return
		}
	}
}
