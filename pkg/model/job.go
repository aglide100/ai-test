package model

import (
	pb_unit_job "github.com/aglide100/ai-test/pb/unit/job"
)

type Job struct {
	PromptA string
	PromptB string
	BlobID string
	Mode string
	Arg string
	ID	string
}

func ProtoToJob(j *pb_unit_job.Job) *Job {
	return &Job{
		PromptA: j.PromptA,
		PromptB: j.PromptB,
		BlobID: j.BlobID,
		Mode: j.Mode,
		Arg: j.Arg,
	}
}