package tests

import (
	"testing"

	"playwallet/internal/domain"
)

// after try, send another kafka msg with same content, should forbid creating another record due to idemptency key
func TestTransferTry(t *testing.T) {
	endpoint, _, teardown := provisionTestApp(t)
	defer teardown(t)
	uid1, uid2 := int64(1), int64(2)
	status, err := makeTransactionReq(endpoint, domain.TransactionReq{
		IdempotencyKey: "hi",
		UserID:         uid1,
		TargetID:       &uid2,
		Amt:            10,
		Type:           domain.Transfer,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("transfer try status: %d\n", status)
}
