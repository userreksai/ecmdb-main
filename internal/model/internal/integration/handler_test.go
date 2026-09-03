//go:build e2e

package integration

import (
	"context"
	"net/http"

	"github.com/userreksai/ecmdb-main/internal/attribute"
	attributemocks "github.com/userreksai/ecmdb-main/internal/attribute/mocks"
	"github.com/userreksai/ecmdb-main/internal/relation"
	relationmocks "github.com/userreksai/ecmdb-main/internal/relation/mocks"
	"github.com/userreksai/ecmdb-main/internal/resource"
	resourcemocks "github.com/userreksai/ecmdb-main/internal/resource/mocks"
	testioc "github.com/userreksai/ecmdb-main/internal/test/ioc"

	"github.com/ecodeclub/ekit/iox"
	"github.com/stretchr/testify/assert"
	"github.com/userreksai/ecmdb-main/internal/model/internal/integration/startup"
	"github.com/userreksai/ecmdb-main/internal/model/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/model/internal/service"
	"github.com/userreksai/ecmdb-main/internal/model/internal/web"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/ginx/test"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"
)

type HandlerTestSuite struct {
	suite.Suite

	dao    dao.ModelDAO
	db     *mongox.Mongo
	server *gin.Engine
	svc    service.Service
}

func (s *HandlerTestSuite) TearDownSuite() {
	_, err := s.db.Collection(dao.ModelCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) TearDownTest() {
	_, err := s.db.Collection(dao.ModelCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerTestSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	attrSvc := attributemocks.NewMockService(ctrl)
	attrSvc.EXPECT().CreateDefaultAttribute(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, s string) (int64, error) {
		return 0, nil
	}).AnyTimes()
	resourceSvc := resourcemocks.NewMockService(ctrl)
	rrSvc := relationmocks.NewMockRelationResourceService(ctrl)
	rmSvc := relationmocks.NewMockRelationModelService(ctrl)

	handler, err := startup.InitHandler(
		&relation.Module{RRSvc: rrSvc, RMSvc: rmSvc},
		&attribute.Module{Svc: attrSvc},
		&resource.Module{Svc: resourceSvc},
		&role.Module{},
		&policy.Module{},
	)
	require.NoError(s.T(), err)
	server := gin.Default()
	handler.PrivateRoutes(server)

	s.db = testioc.InitMongoDB()
	s.dao = dao.NewModelDAO(s.db)
	s.server = server
}

func (s *HandlerTestSuite) TestCreate() {
	testCase := []struct {
		name   string
		req    web.CreateModelReq
		before func(t *testing.T)

		wantCode int
		wantResp test.Result[int64]
	}{
		{
			name: "创建模型",
			req: web.CreateModelReq{
				Name:    "host",
				GroupId: 1,
				UID:     "host",
				Icon:    "icon-host",
			},
			wantCode: 200,
			wantResp: test.Result[int64]{
				Data: 1,
				Msg:  "添加模型成功",
			},
		},
	}

	for _, tc := range testCase {
		s.T().Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost,
				"/api/model/create", iox.NewJSONReader(tc.req))
			req.Header.Set("content-type", "application/json")
			require.NoError(t, err)
			recorder := test.NewJSONResponseRecorder[int64]()
			s.server.ServeHTTP(recorder, req)

			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantResp, recorder.MustScan())
		})
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
