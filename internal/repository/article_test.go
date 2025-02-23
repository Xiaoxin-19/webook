package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webok/internal/domain"
	"webok/internal/repository/dao"
	daomocks "webok/internal/repository/dao/mock"
)

func TestCachedArticleRepository_Sync(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			art: domain.Article{
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id: 123,
				},
			},
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				a := daomocks.NewMockArticleAuthorDAO(ctrl)
				a.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "test title",
					Content:  "test Content",
					AuthorId: 123,
				}).Return(int64(1), nil)

				r := daomocks.NewMockArticleReaderDAO(ctrl)
				r.EXPECT().Upsert(gomock.Any(), dao.Article{
					ID:       1,
					Title:    "test title",
					Content:  "test Content",
					AuthorId: 123,
				}).Return(nil)
				return a, r
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "修改并发表成功",
			art: domain.Article{
				Id:      1,
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id: 123,
				},
			},
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				a := daomocks.NewMockArticleAuthorDAO(ctrl)
				a.EXPECT().UpdateById(gomock.Any(), dao.Article{
					ID:       1,
					Title:    "test title",
					Content:  "test Content",
					AuthorId: 123,
				}).Return(nil)

				r := daomocks.NewMockArticleReaderDAO(ctrl)
				r.EXPECT().Upsert(gomock.Any(), dao.Article{
					ID:       1,
					Title:    "test title",
					Content:  "test Content",
					AuthorId: 123,
				}).Return(nil)
				return a, r
			},
			wantId:  1,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			a, r := tc.mock(ctrl)
			repo := NewCachedArticleRepositoryV1(a, r)
			gotId, gotErr := repo.Sync(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, gotErr)
			assert.Equal(t, tc.wantId, gotId)

		})
	}
}
