package biz

func (uc *WalletUC) CheckBalance(userID int64) (int64, error) {
	return uc.repo.CheckBalance(userID)
}
