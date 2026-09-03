package integration

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/userreksai/ecmdb-main/internal/policy"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/integration/startup"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/repository/dao"
	"github.com/userreksai/ecmdb-main/internal/relation/internal/web"
	"github.com/userreksai/ecmdb-main/internal/role"
	"github.com/userreksai/ecmdb-main/pkg/ginx/test"
	"github.com/userreksai/ecmdb-main/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
)

type HandlerRRTestSuite struct {
	suite.Suite

	dao    dao.RelationResourceDAO
	db     *mongox.Mongo
	server *gin.Engine
}

func (s *HandlerRRTestSuite) TearDownSuite() {
	_, err := s.db.Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerRRTestSuite) TearDownTest() {
	_, err := s.db.Collection(dao.ResourceRelationCollection).DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
	_, err = s.db.Collection("c_id_generator").DeleteMany(context.Background(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *HandlerRRTestSuite) SetupSuite() {
	handler, err := startup.InitRRHandler(&role.Module{}, &policy.Module{})
	require.NoError(s.T(), err)
	server := gin.Default()
	handler.PrivateRoute(server)

	s.db = startup.InitMongoDB()
	s.dao = dao.NewRelationResourceDAO(s.db)
	s.server = server
}

func (s *HandlerRRTestSuite) TestCreate() {
	testCase := []struct {
		name string
		req  web.CreateResourceRelationReq

		wantCode int
		wantResp test.Result[int64]
	}{
		{
			name: "创建成功",
		},
	}

	for _, tc := range testCase {
		s.T().Run(tc.name, func(t *testing.T) {

		})
	}
}

func TestRRHandler(t *testing.T) {
	suite.Run(t, new(HandlerRRTestSuite))
}
