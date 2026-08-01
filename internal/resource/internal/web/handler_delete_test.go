package web

import (
	"errors"
	"testing"

	relationmocks "github.com/Duke1616/ecmdb/internal/relation/mocks"
	resourcemocks "github.com/Duke1616/ecmdb/internal/resource/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlerDeleteResource(t *testing.T) {
	testCases := []struct {
		name           string
		resourceErr    error
		relationErr    error
		expectRelation bool
		expectResult   int64
		expectError    error
	}{
		{
			name:           "delete resource and all of its relations",
			expectRelation: true,
			expectResult:   1,
		},
		{
			name:        "do not delete relations when resource deletion fails",
			resourceErr: errors.New("delete resource failed"),
			expectError: errors.New("delete resource failed"),
		},
		{
			name:           "return relation cleanup error",
			relationErr:    errors.New("delete relations failed"),
			expectRelation: true,
			expectError:    errors.New("delete relations failed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			resourceSvc := resourcemocks.NewMockEncryptedSvc(ctrl)
			relationSvc := relationmocks.NewMockRelationResourceService(ctrl)
			handler := NewHandler(resourceSvc, nil, relationSvc, nil, nil)
			ctx := &gin.Context{}
			const resourceID int64 = 42

			resourceSvc.EXPECT().
				DeleteResource(ctx, resourceID).
				Return(int64(1), tc.resourceErr)
			if tc.expectRelation {
				relationSvc.EXPECT().
					DeleteRelationsByResourceID(ctx, resourceID).
					Return(int64(3), tc.relationErr)
			}

			result, err := handler.DeleteResource(ctx, DeleteResourceReq{Id: resourceID})

			if tc.expectError != nil {
				assert.EqualError(t, err, tc.expectError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, result.Data)
		})
	}
}
