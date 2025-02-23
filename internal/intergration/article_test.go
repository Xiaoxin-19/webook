package intergration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"webok/internal/intergration/startup"
	"webok/internal/repository/dao"
	ijwt "webok/internal/web/jwt"
)

type Result[T any] struct {
	Code uint16 `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (s *ArticleHandlerSuite) SetupSuite() {
	s.db = startup.InitDB()
	hdl := startup.InitArticleHandler()
	s.server = gin.Default()
	s.server.Use(func(c *gin.Context) {
		c.Set("user", ijwt.TokenClaims{Uid: 123})
	})
	hdl.RegisterRoutes(s.server)
}

func (s *ArticleHandlerSuite) TearDownTest() {
	// 清空数据,并重置自增ID
	s.db.Exec("ALTER SEQUENCE articles_id_seq RESTART WITH 1")
	s.db.Exec("TRUNCATE TABLE articles")

}

func (s *ArticleHandlerSuite) Test_edit() {
	t := s.T()
	testCases := []struct {
		name string
		art  Article

		before     func(t *testing.T)
		after      func(t *testing.T)
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建文章",
			art:  Article{Title: "test1", Content: "test1"},
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				a := dao.Article{}
				s.db.Where("author_id = ?", 123).First(&a)
				assert.Equal(t, "test1", a.Title)
				assert.Equal(t, "test1", a.Content)
				assert.True(t, a.ID > 0)
				assert.Equal(t, int64(123), a.AuthorId)
				assert.True(t, a.Ctime > 0)
				assert.True(t, a.Utime > 0)
			},
			wantCode:   http.StatusOK,
			wantResult: Result[int64]{Data: 1},
		},
		{
			name: "修改文章",
			art:  Article{Id: 1, Title: "test1", Content: "test2"},
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{Title: "test1", Content: "test1", AuthorId: 123, Ctime: 789, Utime: 789}).Error
				assert.Nil(t, err)
			},
			after: func(t *testing.T) {
				a := dao.Article{}
				s.db.Where("author_id = ?", 123).First(&a)
				assert.Equal(t, "test1", a.Title)
				assert.Equal(t, "test2", a.Content)
				assert.Equal(t, int64(1), a.ID)
				assert.Equal(t, int64(123), a.AuthorId)
				assert.Equal(t, int64(789), a.Ctime)
				assert.True(t, a.Utime != 789)
			},
			wantCode:   http.StatusOK,
			wantResult: Result[int64]{Data: 1},
		},
		{
			name: "修改别人的文章",
			art:  Article{Id: 1, Title: "test1", Content: "test2"},
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{Title: "test1", Content: "test1", AuthorId: 234, Ctime: 789, Utime: 789}).Error
				assert.Nil(t, err)
			},
			after: func(t *testing.T) {
				a := dao.Article{}
				s.db.Where("author_id = ?", 234).First(&a)
				assert.Equal(t, "test1", a.Title)
				assert.Equal(t, "test1", a.Content)
				assert.Equal(t, int64(1), a.ID)
				assert.Equal(t, int64(234), a.AuthorId)
				assert.Equal(t, int64(789), a.Ctime)
				assert.Equal(t, int64(789), a.Utime)
			},
			wantCode:   http.StatusOK,
			wantResult: Result[int64]{Data: 0, Msg: "系统错误", Code: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			// 不要全部测试用例共用一个ctrl
			ctrl := gomock.NewController(t)
			defer func() {
				ctrl.Finish()
				tc.after(t)
				s.TearDownTest()
			}()

			reqBody, err := json.Marshal(tc.art)
			assert.Nil(t, err)
			req := httptest.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.Nil(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResult, res)
		})
	}
}

func TestArticleHandlerSuite(t *testing.T) {
	suite.Run(t, new(ArticleHandlerSuite))
}
