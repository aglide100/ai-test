package server

import (
	"context"
	"flag"
	"strings"
	"time"

	pb_svc_fixer "github.com/aglide100/ai-test/pb/svc/fixer"
	"go.uber.org/zap"

	pb_unit_response "github.com/aglide100/ai-test/pb/unit/response"

	pb_unit_blob "github.com/aglide100/ai-test/pb/unit/blob"
	"github.com/aglide100/ai-test/pkg/cache"
	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/aglide100/ai-test/pkg/queue"
	"github.com/google/uuid"
)

var (
	modes = flag.String("modes", "", "modes")
	isReturnString = flag.Bool("isReturnString", false, "return to string or bytes")
)

type FixerSrv struct {
	pb_svc_fixer.FixerServiceServer
	taskAllocator   *queue.TaskAllocator
	resultCache     *cache.Cache
	blobCache       *cache.BlobCache
	readableRequest chan *model.RequestData
	doneJob         chan string
	token           string
	returnString    bool
	timeOutTime     int
	modes 			[]string
}

func NewFixerServiceServer(taskAllocator *queue.TaskAllocator, token string, doneJob chan string, readableRequest chan *model.RequestData, resultCache *cache.Cache, blobCache *cache.BlobCache,
	timeOutTime int) *FixerSrv {

	supportModes := strings.Split(*modes, ",")
	logger.Info("supported modes", zap.Any("modes", supportModes))
	
	return &FixerSrv{
		taskAllocator:   taskAllocator,
		readableRequest: readableRequest,
		token:           token,
		doneJob:         doneJob,
		resultCache:     resultCache,
		blobCache:       blobCache,
		returnString:    *isReturnString,
		timeOutTime:     timeOutTime,
		modes: supportModes,
	}
}

func (s FixerSrv) CheckClients(ctx context.Context, in *pb_svc_fixer.CheckClientsReq) (*pb_svc_fixer.CheckClientsRes, error) {
	// TODO;
	clients := []string{}

	return &pb_svc_fixer.CheckClientsRes{
		Res: clients,
	}, nil
}

func (s FixerSrv) GetBlob(ctx context.Context, in *pb_svc_fixer.GetBlobReq) (*pb_svc_fixer.GetBlobRes, error) {
	if in.Token == "" || in.Token != s.token {
		logger.Info("not found", zap.String("token", in.Token), zap.Any("key", in.Key))

		return &pb_svc_fixer.GetBlobRes{}, nil
	}

	res, found := s.blobCache.Get(in.Key)
	if !found {
		logger.Info("can't find blob", zap.String("token", in.Token), zap.Any("key", in.Key))
		return &pb_svc_fixer.GetBlobRes{}, nil
	}

	s.blobCache.Delete(in.Key)

	data := res

	return &pb_svc_fixer.GetBlobRes{
		Blob: &pb_unit_blob.Blob{
			Data: data,
		},
	}, nil
}

func (s FixerSrv) SendBlob(ctx context.Context, in *pb_svc_fixer.SendBlobReq) (*pb_svc_fixer.SendBlobRes, error) {
	if in.Token == "" || in.Token != s.token {
		logger.Info("not found", zap.String("token", in.Token))

		return &pb_svc_fixer.SendBlobRes{}, nil
	}

	blobID := uuid.New().String()

	logger.Info("set blob", zap.Any("id", blobID), zap.Any("byte", len(in.Blob.Data)))
	s.blobCache.Set(blobID, in.Blob.Data, true)

	return &pb_svc_fixer.SendBlobRes{
		BlobID: blobID,
	}, nil
}

func (s FixerSrv) GetResult(ctx context.Context, in *pb_svc_fixer.GetResultReq) (*pb_svc_fixer.GetResultRes, error) {
	if in.Auth == nil || in.Auth.Token != s.token {
		return &pb_svc_fixer.GetResultRes{
			Error: &pb_svc_fixer.Error{
				Msg: "",
			},
		}, nil
	}

	res, found := s.resultCache.Get(in.JobId)
	if !found {
		msg := s.taskAllocator.Check(in.JobId)
		
		return &pb_svc_fixer.GetResultRes{
			Error: &pb_svc_fixer.Error{
				Msg: msg,
			},
		}, nil
	}

	data := res.([]byte)
	return &pb_svc_fixer.GetResultRes{
		Res: &pb_unit_response.Response{
			Binary: data,
		},
	}, nil
}

func contains(s []string, substr string) bool {
    for _, v := range s {
        if v == substr {
            return true
        }
    }

    return false
}

func (s FixerSrv) MakingNewJob(ctx context.Context, in *pb_svc_fixer.MakingNewJobReq) (*pb_svc_fixer.MakingNewJobRes, error) {
	if in.Auth == nil || in.Auth.Token != s.token {
		return &pb_svc_fixer.MakingNewJobRes{
			Error: &pb_svc_fixer.Error{
				Msg: "",
			},
		}, nil
	}

	if in.Job == nil {
		return &pb_svc_fixer.MakingNewJobRes{
			Error: &pb_svc_fixer.Error{
				Msg: "job is nil",
			},
		}, nil
	}


	job := model.ProtoToJob(in.Job)
	if job.Mode == "" {
		return &pb_svc_fixer.MakingNewJobRes{
			Error: &pb_svc_fixer.Error{
				Msg: "mode is null",
			},
		}, nil
	}

	if !contains(s.modes, in.Job.Mode) {
		return &pb_svc_fixer.MakingNewJobRes{
			Error: &pb_svc_fixer.Error{
				Msg: "can't find supported mode",
			},
		}, nil
	}

	jobId := uuid.New().String()

	job.ID = jobId

	s.taskAllocator.AddInWait(job)
	// logger.Info("add job", zap.Any("waiting", s.taskAllocator.LenWaiting()))
	if in.IsWait {
		responseChan := make(chan []byte, 1)
		requestData := &model.RequestData{
			ResponseChan: responseChan,
			Data:         jobId,
		}

		s.readableRequest <- requestData

		timeout := time.After(time.Duration(s.timeOutTime) * time.Second)

		select {
		case response := <-responseChan:
			key := string(response[:])
			key = strings.Replace(key, "\"", "", -1)
			var data []byte

			s.blobCache.Delete(jobId)
			if len(key) == 36 {
				res, found := s.blobCache.Get(key)
				if !found {
					logger.Info("error", zap.Any("key", key))
					return &pb_svc_fixer.MakingNewJobRes{
						Res: &pb_unit_response.Response{},
					}, nil
				}

				s.blobCache.Delete(key)
				data = res
			} else {
				data = response
			}

			if s.returnString {
				return &pb_svc_fixer.MakingNewJobRes{
					Res: &pb_unit_response.Response{
						Data: string(data[:]),
					},
				}, nil
			} else {
				return &pb_svc_fixer.MakingNewJobRes{
					Res: &pb_unit_response.Response{
						Binary: []byte(data),
					},
				}, nil
			}

		case <-timeout:
			close(responseChan)
			return &pb_svc_fixer.MakingNewJobRes{
				JobId: jobId,
				Res: &pb_unit_response.Response{
					Msg: "timeout",
				},
			}, nil
		}
	} else {
		requestData := &model.RequestData{}
		s.readableRequest <- requestData
	}

	return &pb_svc_fixer.MakingNewJobRes{
		JobId: jobId,
		Res: &pb_unit_response.Response{
			Msg: "later check please",
		},
	}, nil
}
