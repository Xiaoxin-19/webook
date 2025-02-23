package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webok/internal/domain"
	"webok/internal/repository"
	repomocks "webok/internal/repository/mock"
	"webok/pkg/logger"
)

func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发布成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(int64(1), nil)

				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(nil)

				return author, reader
			},
			art: domain.Article{
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id:   123,
					Name: "name",
				},
			},
			wantErr: nil,
			wantId:  1,
		},
		{
			name: "新建并发表失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(int64(1), nil)

				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(errors.New("publish error")).MaxTimes(3)

				return author, reader
			},
			art: domain.Article{
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id:   123,
					Name: "name",
				},
			},
			wantErr: errors.New("publish error"),
			wantId:  0,
		},
		{
			name: "修改并新发布失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(nil)

				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(errors.New("publish error")).Times(3)

				return author, reader
			},
			art: domain.Article{
				Id:      1,
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id:   123,
					Name: "name",
				},
			},
			wantErr: errors.New("publish error"),
			wantId:  0,
		},
		{
			name: "修改并新发布失败,重试成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(nil)

				reader := repomocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(errors.New("publish error")).Times(2)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(nil)

				return author, reader
			},
			art: domain.Article{
				Id:      1,
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id:   123,
					Name: "name",
				},
			},
			wantErr: nil,
			wantId:  1,
		},

		{
			name: "保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				author := repomocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "test title",
					Content: "test Content",
					Author: domain.Author{
						Id:   123,
						Name: "name",
					},
				}).Return(int64(0), errors.New("create error"))

				return author, nil
			},
			art: domain.Article{
				Title:   "test title",
				Content: "test Content",
				Author: domain.Author{
					Id:   123,
					Name: "name",
				},
			},
			wantErr: errors.New("create error"),
			wantId:  0,
		},
	}
	l := logger.NewNopLogger()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authorRepo, readerRepo := tc.mock(ctrl)
			svc := NewArticleServiceV1(readerRepo, authorRepo, l)
			gotId, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, gotId)
		})
	}
}
