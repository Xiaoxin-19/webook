package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
	"webok/internal/domain"
	"webok/internal/repository/cache"
	"webok/internal/repository/dao"
)

//go:generate mockgen -source=article.go -package=repomocks -destination=./mock/article.mock.go
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, articleId int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	preCache(ctx context.Context, arts []domain.Article)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDAO
	userRepo UserRepository

	// V1
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO

	//V2
	db *gorm.DB

	cache cache.ArticleCache
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.ToDoMain(dao.Article(art))
	// 查询Author Name
	user, err := c.userRepo.FindById(ctx, res.Author.Id)
	if err != nil {
		// log
		return res, err
	}
	res.Author.Name = user.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			// log
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	article, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.ToDoMain(article)
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.ToEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	go func() {
		// 可以灵活设置过期时间
		user, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			//log
		}
		art.Author.Name = user.Nickname
		art.Author.Id = user.Id
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			//log
		}
	}()
	return id, err
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

func NewCachedArticleRepository(d dao.ArticleDAO, db *gorm.DB, c cache.ArticleCache, ur UserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      d,
		db:       db,
		cache:    c,
		userRepo: ur,
	}
}

func NewCachedArticleRepositoryV1(a dao.ArticleAuthorDAO, r dao.ArticleReaderDAO) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: a,
		readerDAO: r,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.ToEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := c.dao.UpdateById(ctx, c.ToEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			// 记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) ToEntity(article domain.Article) dao.Article {
	return dao.Article{
		ID:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) ToDoMain(article dao.Article) domain.Article {
	return domain.Article{
		Id:      article.ID,
		Title:   article.Title,
		Content: article.Content,
		Author: domain.Author{
			Id: article.AuthorId,
		},
		Status: domain.ArticleStatus(article.Status),
		Ctime:  article.Ctime,
		Utime:  article.Utime,
	}
}
func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, articleId int64) error {

	err := c.dao.SyncStatus(ctx, uid, articleId, domain.ArticleStatusPrivate)
	if err == nil {
		er := c.cache.DelFirstPage(ctx, uid)
		if er != nil {
			// 记录日志
		}
	}
	return err
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if limit <= 100 && offset == 0 {
		// 从缓存中读取
		articles, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return articles, nil
		} else {
			if !errors.Is(err, cache.ErrKeyNotExist) {
				// 记录日志
				return nil, err
			}
		}
	}
	articles, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	var result []domain.Article
	for _, article := range articles {
		result = append(result, c.ToDoMain(article))
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = c.cache.SetFirstPage(ctx, uid, result)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, result)
	}()
	return result, nil
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
		}
	}
}
