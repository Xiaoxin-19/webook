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

type WithdrawArticleReq struct {
	Id int64 `json:"id"`
}

type ArticleVO struct {
	ID         int64  `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Abstract   string `json:"abstract,omitempty"`
	Content    string `json:"content,omitempty"`
	AuthorId   int64  `json:"author_id,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Ctime      int64  `json:"ctime,omitempty"`
	Utime      int64  `json:"utime,omitempty"`
}
