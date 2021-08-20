package ginext

const DefaultQueryLimit = 20

type PaginationForm struct {
	Keyword *string `json:"keyword"`
	Pos     *int    `json:"pos"`
	Page    *int    `json:"page"`
	Limit   *int    `json:"limit"`
}

func (p PaginationForm) GetPos() int {
	if p.Page != nil {
		return *p.Page * p.GetLimit()
	}
	if p.Pos != nil {
		return *p.Pos
	}
	return 0
}

func (p PaginationForm) GetLimit() int {
	if p.Limit != nil {
		return *p.Limit
	}
	return DefaultQueryLimit
}

func (p PaginationForm) GetKeyword() string {
	if p.Keyword != nil {
		return *p.Keyword
	}
	return ""
}

type PaginationResult struct {
	Pos        int `json:"pos"`
	Limit      int `json:"limit"`
	TotalCount int `json:"totalCount"`
}
