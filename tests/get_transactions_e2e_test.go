package tests

import (
	"testing"

	"playwallet/internal/domain"
)

func TestGetTransactions(t *testing.T) {
	endpoint, db, teardown := provisionTestApp(t)
	defer teardown(t)
	actions := []action{
		{userID: 1, actType: deposit, amt: 100},
		{userID: 1, actType: withdraw, amt: 50},              // user1: 50
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
		{userID: 1, actType: transfer, targetID: 2, amt: 20}, // user1: 50-20=30, user2: 60+20=80
	}
	addTestData(t, db, actions...)
	uid := int64(1)
	_, res, err := makeGetTransactionsReq(endpoint, uid, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != len(actions) {
		t.Fatalf("invalid response length, expect: %d, actual: %d\n", len(actions), res.Total)
	}
	if len(res.Records) != 10 {
		t.Fatalf("if pageopt not pass, should return 10 records by default, instead: %d\n", len(res.Records))
	}

	// first page
	_, res, err = makeGetTransactionsReq(endpoint, uid, &domain.PageOpt{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != len(actions) {
		t.Fatalf("invalid response length, expect: %d, actual: %d\n", len(actions), res.Total)
	}
	if len(res.Records) != 10 {
		t.Fatalf("if should return 10 records for page 1, instead: %d\n", len(res.Records))
	}

	// second page
	_, res, err = makeGetTransactionsReq(endpoint, uid, &domain.PageOpt{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != len(actions) {
		t.Fatalf("invalid response length, expect: %d, actual: %d\n", len(actions), res.Total)
	}
	if len(res.Records) != 1 {
		t.Fatalf("if should return 1 record for page 2, instead: %d\n", len(res.Records))
	}
}
