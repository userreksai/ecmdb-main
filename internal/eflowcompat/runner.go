package eflowcompat

import (
	"context"
	"fmt"

	runnerv1 "github.com/userreksai/ecmdb-main/api/proto/gen/ecmdb/runner/v1"
	"github.com/userreksai/ecmdb-main/internal/codebook"
	"github.com/userreksai/ecmdb-main/internal/runner"
	"github.com/userreksai/ecmdb-main/pkg/cryptox"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRunnersPerCodebook = 1000

// RunnerServer 实现 ecmdb.runner.v1.RunnerService。
//
// eflow 以数字 ID 引用代码模板（codebook_id），ECMDB 执行单元内部以 codebook_uid 关联，
// 这里负责两者的换算；ResolveExecution 返回解密后的变量，仅供内网 ecmdb-task 调用（受 gRPC JWT 保护）。
type RunnerServer struct {
	runnerv1.UnimplementedRunnerServiceServer
	runners   runner.Service
	codebooks codebook.Service
	crypto    cryptox.Crypto
}

// NewRunnerServer 创建执行单元服务；crypto 为执行单元变量的加解密器（CryptoRegistry.Runner）。
func NewRunnerServer(runners runner.Service, codebooks codebook.Service, crypto cryptox.Crypto) *RunnerServer {
	return &RunnerServer{runners: runners, codebooks: codebooks, crypto: crypto}
}

func (s *RunnerServer) FindRunnerByID(ctx context.Context, req *runnerv1.FindRunnerByIDRequest) (*runnerv1.FindRunnerByIDResponse, error) {
	r, cb, err := s.load(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &runnerv1.FindRunnerByIDResponse{Runner: toProtoRunner(r, cb.Id)}, nil
}

func (s *RunnerServer) ListRunnersByCodebookID(ctx context.Context,
	req *runnerv1.ListRunnersByCodebookIDRequest) (*runnerv1.ListRunnersByCodebookIDResponse, error) {
	cb, err := s.codebooks.DetailCodebook(ctx, req.GetCodebookId())
	if err != nil || cb.Id <= 0 {
		return nil, status.Errorf(codes.NotFound, "代码模板 %d 不存在", req.GetCodebookId())
	}
	runners, _, err := s.runners.ListByCodebookUid(ctx, 0, maxRunnersPerCodebook, cb.Identifier, "", "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*runnerv1.Runner, 0, len(runners))
	for _, r := range runners {
		out = append(out, toProtoRunner(r, cb.Id))
	}
	return &runnerv1.ListRunnersByCodebookIDResponse{Runners: out}, nil
}

func (s *RunnerServer) ResolveExecution(ctx context.Context,
	req *runnerv1.ResolveExecutionRequest) (*runnerv1.ResolveExecutionResponse, error) {
	r, cb, err := s.load(ctx, req.GetRunnerId())
	if err != nil {
		return nil, err
	}
	vars := make([]*runnerv1.Variable, 0, len(r.Variables))
	for _, v := range r.Variables {
		value := v.Value
		if v.Secret {
			if value, err = s.crypto.Decrypt(v.Value); err != nil {
				return nil, status.Errorf(codes.Internal, "解密执行单元 %d 的变量 %s 失败", r.Id, v.Key)
			}
		}
		vars = append(vars, &runnerv1.Variable{Key: v.Key, Value: value, Secret: v.Secret})
	}
	return &runnerv1.ResolveExecutionResponse{Execution: &runnerv1.RunnerExecution{
		Runner: toProtoRunner(r, cb.Id), Code: cb.Code, Language: cb.Language, Variables: vars,
	}}, nil
}

func (s *RunnerServer) load(ctx context.Context, runnerID int64) (runner.Runner, codebook.Codebook, error) {
	r, err := s.runners.FindById(ctx, runnerID)
	if err != nil || r.Id <= 0 {
		return runner.Runner{}, codebook.Codebook{}, status.Errorf(codes.NotFound, "执行单元 %d 不存在", runnerID)
	}
	cb, err := s.codebooks.FindByUid(ctx, r.CodebookUid)
	if err != nil || cb.Id <= 0 {
		return runner.Runner{}, codebook.Codebook{}, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("执行单元 %d 绑定的代码模板 %s 不存在", runnerID, r.CodebookUid))
	}
	return r, cb, nil
}

// toProtoRunner 不包含变量：变量只在 ResolveExecution 中返回。
func toProtoRunner(r runner.Runner, codebookID int64) *runnerv1.Runner {
	return &runnerv1.Runner{
		Id: r.Id, Name: r.Name, CodebookId: codebookID, CodebookUid: r.CodebookUid,
		Kind: string(r.Kind), Target: r.Target, Handler: r.Handler, Tags: r.Tags, Desc: r.Desc,
	}
}
