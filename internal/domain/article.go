package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

type ArticleStatus uint8

func (as ArticleStatus) ToUint8() uint8 {
	return uint8(as)
}

const (
	// ArticleStatusUnknown 未知状态
	ArticleStatusUnknown ArticleStatus = iota
	// ArticleStatusUnpublished 未发布
	ArticleStatusUnpublished
	// ArticleStatusPublished 已发布
	ArticleStatusPublished
	// ArticleStatusPrivate 私有 不可见
	ArticleStatusPrivate
)

type Author struct {
	Id   int64
	Name string
}
