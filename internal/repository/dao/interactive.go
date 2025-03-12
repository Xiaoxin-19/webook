package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Interactive struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement;comment:互动ID"`
	BizId      int64  `gorm:"column:biz_id;comment:业务ID;uniqueIndex:biz_type_id"`
	Biz        string `gorm:"column:biz;type:varchar(128);comment:业务类型;uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

type UserLikeBiz struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement;comment:互动ID"`
	UserId int64  `gorm:"column:uid;comment:用户ID;uniqueIndex:like_uid_biz_type_id"`
	BizId  int64  `gorm:"column:biz_id;comment:业务ID;uniqueIndex:like_uid_biz_type_id"`
	Biz    string `gorm:"column:biz;type:varchar(128);comment:业务类型;uniqueIndex:like_uid_biz_type_id"`
	Status uint8
	Ctime  int64
	Utime  int64
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这边还是保留了了唯一索引
	Uid   int64  `gorm:"column:uid;uniqueIndex:collect_uid_biz_type_id"`
	BizId int64  `gorm:"column:biz_id;uniqueIndex:collect_uid_biz_type_id"`
	Biz   string `gorm:"column:biz;type:varchar(128);uniqueIndex:collect_uid_biz_type_id"`
	// 收藏夹的ID
	// 收藏夹ID本身有索引
	Cid   int64 `gorm:"index"`
	Utime int64
	Ctime int64
}

//go:generate mockgen -source=interactive.go -package=daomocks -destination=./mock/interactive.mock.go
type InteractiveDao interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	IncrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	DecrLickCnt(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetLikedInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectionInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
}

type InteractiveGORMDAO struct {
	db *gorm.DB
}

func (i *InteractiveGORMDAO) GetCollectionInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var collection UserCollectionBiz
	err := i.db.WithContext(ctx).Where("uid = ? AND biz = ? AND biz_id = ?", uid, biz, id).First(&collection).Error
	if err != nil {
		return UserCollectionBiz{}, err
	}
	return collection, nil
}

func (i *InteractiveGORMDAO) GetLikedInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {

	var like UserLikeBiz
	err := i.db.WithContext(ctx).Where("uid = ? AND biz = ? AND biz_id = ? AND status = ?", uid, biz, id, 1).First(&like).Error
	if err != nil {
		return UserLikeBiz{}, err
	}
	return like, nil
}

func (i *InteractiveGORMDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var intr Interactive
	err := i.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, id).First(&intr).Error
	return intr, err
}

func (i *InteractiveGORMDAO) InsertCollectionBiz(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	now := time.Now().UnixMilli()
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserCollectionBiz{}).Create(&UserCollectionBiz{
			Uid:   uid,
			BizId: id,
			Biz:   biz,
			Cid:   cid,
			Ctime: now,
			Utime: now,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "biz_id"}, {Name: "biz"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("public.interactives.collect_cnt + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			BizId:      id,
			Biz:        biz,
			Utime:      now,
			Ctime:      now,
			CollectCnt: 1,
		}).Error
	})
}

func (i *InteractiveGORMDAO) IncrLickCnt(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()

	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "uid"}, {Name: "biz_id"}, {Name: "biz"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			UserId: uid,
			BizId:  id,
			Biz:    biz,
			Ctime:  now,
			Utime:  now,
			Status: 1,
		}).Error

		if err != nil {
			return err
		}

		err = tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "biz_id"}, {Name: "biz"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("webook.public.interactives.like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			BizId:   id,
			Biz:     biz,
			Utime:   now,
			Ctime:   now,
			LikeCnt: 1,
		}).Error
		return err
	})
}

func (i *InteractiveGORMDAO) DecrLickCnt(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Model(&Interactive{}).Where("biz =? AND biz_id=?", biz, id).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("public.interactives.like_cnt - 1"),
				"utime":    now,
			}).Error
		if err != nil {
			return err
		}

		// 软删除用户点赞信息
		err = tx.WithContext(ctx).Model(&UserLikeBiz{}).Where("uid=? AND biz =? AND biz_id=?", uid, biz, id).Updates(map[string]any{
			"status": 0,
			"utime":  now,
		}).Error
		return err
	})
}

func (i *InteractiveGORMDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	now := time.Now().UnixMilli()
	err := i.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}, {Name: "biz"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("public.interactives.read_cnt + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		BizId:   id,
		Biz:     biz,
		Utime:   now,
		Ctime:   now,
		ReadCnt: 1,
	}).Error
	return err
}

func NewInteractiveGORMDAO(db *gorm.DB) InteractiveDao {
	return &InteractiveGORMDAO{db: db}
}
