package tests

import (
	"testing"
)

// we just mock some data in db to test the check balance calculation
// since we are mocking data, we would make sure not to produce invalid data eg: negative balance
func TestCheckBalance(t *testing.T) {
	cases := []struct {
		desc           string
		actions        []action
		expectBalances [2]int64 // expected balance for user 1 and 2
	}{
		{
			desc:           "no action",
			actions:        []action{},
			expectBalances: [2]int64{0, 0},
		},
		{
			desc: "deposit",
			actions: []action{
				{userID: 1, actType: deposit, amt: 10},
				{userID: 1, actType: deposit, amt: 15},
			},
			expectBalances: [2]int64{25, 0},
		},
		{
			desc: "deposit & withdraw",
			actions: []action{
				{userID: 1, actType: deposit, amt: 10},
				{userID: 1, actType: deposit, amt: 15},
				{userID: 1, actType: withdraw, amt: 5},
			},
			expectBalances: [2]int64{20, 0},
		},
		{
			desc: "transfer",
			actions: []action{
				{userID: 1, actType: deposit, amt: 10},
				{userID: 1, actType: deposit, amt: 15},
				{userID: 1, actType: withdraw, amt: 5},
				{userID: 1, actType: transfer, targetID: 2, amt: 2},
			},
			expectBalances: [2]int64{18, 2},
		},
		{
			desc: "pending transfer(freeze)",
			actions: []action{
				{userID: 1, actType: deposit, amt: 100},
				{userID: 1, actType: transfer, targetID: 2, amt: 2},
				{userID: 1, actType: transfer, targetID: 2, amt: 2},
				{userID: 1, actType: freeze, targetID: 2, amt: 2},
			},
			expectBalances: [2]int64{94, 4},
		},
		{
			desc: "bi-direction transfer",
			actions: []action{
				{userID: 1, actType: deposit, amt: 100},
				{userID: 1, actType: withdraw, amt: 50}, // user1: 50
				{userID: 2, actType: deposit, amt: 100},
				{userID: 2, actType: withdraw, amt: 40},              // user2: 60
				{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
				{userID: 2, actType: transfer, targetID: 1, amt: 25}, // user1:30+25=55, user2: 80-25 = 55
				{userID: 1, actType: freeze, targetID: 2, amt: 10},   // user1: 55-10=45, user2: 55
				{userID: 2, actType: freeze, targetID: 1, amt: 5},    // user1: 45, user2: 55-5=50,
			},
			expectBalances: [2]int64{45, 50},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			endpoint, db, teardown := provisionTestApp(t)
			defer teardown(t)
			if err := addTestData(t, db, c.actions...); err != nil {
				t.Fatal(err)
			}
			uids := []int64{1, 2}
			for i, uid := range uids {
				_, bInfo, err := makeCheckBalanceReq(endpoint, uid)
				if err != nil {
					t.Fatal(err)
				}
				if !bInfo.IsValid() {
					t.Fatalf("invalid balance info: %+v", bInfo)
				}
				if bInfo.UserID != uid {
					t.Fatalf("balance userid mismatch, expect: %d, actual: %d\n", uid, bInfo.UserID)
				}
				if bInfo.AvailableBalance != c.expectBalances[i] {
					t.Fatalf("unexpected AvailableBalance result: %d, expect: %d, bInfo = %+v\n",
						bInfo.AvailableBalance, c.expectBalances[i], bInfo)
				}
			}
		})
	}
}
