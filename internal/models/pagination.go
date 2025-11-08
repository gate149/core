package models

type Pagination struct {
	Page  int32 `json:"page"`
	Total int32 `json:"total"`
}

func Total(count int32, pageSize int32) int32 {
	if count%pageSize == 0 {
		return count / pageSize
	}
	return count/pageSize + 1
}
