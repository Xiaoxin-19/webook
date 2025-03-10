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
	Biz        string `gorm:"column:biz;comment:业务类型;uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

//go:generate mockgen -source=interactive.go -package=daomocks -destination=./mock/interactive.mock.go
type InteractiveDao interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
}

type InteractiveGORMDAO struct {
	db *gorm.DB
}

func (i *InteractiveGORMDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	now := time.Now().UnixMilli()
	err := i.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("webook.public.interactives.read_cnt + 1"),
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
