package repository

import (
	"context"
	"gorm.io/gorm"
	"webok/internal/domain"
	"webok/internal/repository/dao"
)

//go:generate mockgen -source=article.go -package=repomocks -destination=./mock/article.mock.go
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDAO

	// V1
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO

	//V2
	db *gorm.DB
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.ToEntity(art))
}

// SyncV1 Repo不使用事务，保证数据一致性
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  int64
		err error
	)
	entity := c.ToEntity(art)
	id = entity.ID
	if id > 0 {
		err = c.authorDAO.UpdateById(ctx, entity)
	} else {
		id, err = c.authorDAO.Create(ctx, entity)
	}

	if err != nil {
		return 0, err
	}
	entity.ID = id
	return id, c.readerDAO.Upsert(ctx, entity)
}

// SyncV2 Repo层通过事务保证数据一致性
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止之后的业务panic
	defer tx.Rollback()

	aDao := dao.NewArticleAuthorGORMDAO(tx)
	rDao := dao.NewArticleReaderGORMDAO(tx)

	entity := c.ToEntity(art)
	var (
		id  int64
		err error
	)
	id = entity.ID
	if id == 0 {
		id, err = aDao.Create(ctx, entity)
	} else {
		err = aDao.UpdateById(ctx, entity)
	}
	if err != nil {
		return 0, err
	}

	entity.ID = id
	err = rDao.UpsertV2(ctx, dao.PublishedArticle{
		ID:       entity.ID,
		Title:    entity.Title,
		Content:  entity.Content,
		AuthorId: entity.AuthorId,
	})
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return entity.ID, nil
}

func NewCachedArticleRepository(d dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{dao: d}
}

func NewCachedArticleRepositoryV1(a dao.ArticleAuthorDAO, r dao.ArticleReaderDAO) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: a,
		readerDAO: r,
	}
}
func NewCachedArticleRepositoryV2(a dao.ArticleAuthorDAO, r dao.ArticleReaderDAO, db *gorm.DB) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: a,
		readerDAO: r,
		db:        db,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return c.dao.Insert(ctx, c.ToEntity(article))
}

func (c *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return c.dao.UpdateById(ctx, c.ToEntity(article))
}

func (c *CachedArticleRepository) ToEntity(article domain.Article) dao.Article {
	return dao.Article{
		ID:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}
