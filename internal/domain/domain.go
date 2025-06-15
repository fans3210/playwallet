package domain

// 1 based index
type PageOpt struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func (p PageOpt) IsValid() bool {
	return p.Page >= 1 && p.PerPage >= 1
}
