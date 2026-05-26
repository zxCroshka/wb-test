package domain

type RankItem struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}
