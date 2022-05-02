package job

import (
	"github.com/bingxindan/bxd_go_lib/logger"
	"github.com/bingxindan/bxd_go_lib/tools/flagutil"
)

type Job struct {
	// 任务名称
	Name string
	Task TaskFunc
}

type TaskFunc func() error

type CmdParser interface {
	// 解析命令行参数，并选择对应的job任务
	JobArgParse(jobs map[string]Job) (selectedJobs []Job, err error)
}

type defaultCmdParse struct {
}

func (p *defaultCmdParse) JobArgParse(jobs map[string]Job) (selectedJobs []Job, err error) {
	cmdArg := *flagutil.GetExtendedopt()
	if cmdArg == "" {
		return nil, logger.NewError("请使用参数 -extended 选择任务, 如：-extended testJob")
	}

	job, ok := jobs[cmdArg]
	if !ok {
		return nil, logger.NewError("[ " + cmdArg + " ]任务未定义")
	}

	selectedJobs = make([]Job, 0, 1)
	selectedJobs = append(selectedJobs, job)

	return selectedJobs, nil
}

func (j *Job) Do() error {
	if j.Task == nil {
		return logger.NewError("方法未定义")
	}
	return j.Task()
}
