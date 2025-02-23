package web

type EditArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PublishArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
