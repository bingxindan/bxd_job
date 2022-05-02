package job

import (
	"context"
	"github.com/bingxindan/bxd_go_lib/logger"
	"github.com/bingxindan/bxd_go_lib/tools/confutil"
	"github.com/bingxindan/bxd_job/bootstrap"
	"github.com/spf13/cast"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	*bootstrap.FuncSetter
	Opts Options
	Jobs map[string]Job
	// 退出
	exit      chan struct{}
	cmdParser CmdParser
}

func init() {
	confutil.InitConfig()
}

func NewJobServer(options ...OptionFunc) *Server {
	opts := DefaultOptions()

	for _, o := range options {
		o(&opts)
	}

	srv := &Server{
		Opts:       opts,
		FuncSetter: bootstrap.NewFuncSetter(),
		Jobs:       make(map[string]Job),
		exit:       make(chan struct{}),
		cmdParser:  opts.cmdParser,
	}

	return srv
}

func (s *Server) AddJobs(jobs map[string]Job) error {
	if len(jobs) == 0 {
		return logger.NewError("请注入任务")
	}
	for key, j := range jobs {
		s.Jobs[key] = j
	}

	return nil
}

func (s *Server) Start() (err error) {
	defer recoverProc()
	if err = s.RunBeforeServerStartFunc(); err != nil {
		return nil
	}

	go s.dealExitSignal()

	return nil
}

func (s *Server) doJob() {
	tag := "doJob"

	defer s.Stop()
	defer processMark(tag, "任务主入口")()

	jobSelected, err := s.cmdParser.JobArgParse(s.Jobs)
	if err != nil {
		logger.Ex(context.Background(), tag, "解析命令错误:%v", err)
		return
	}

	wg := &sync.WaitGroup{}
	for _, myjob := range jobSelected {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			defer processMark(tag, "任务:"+job.Name)()
			if err := job.Do(); err != nil {
				logger.Ex(context.Background(), tag, "[%v] error: %v", job.Name, err)
			}
		}(myjob)
	}
	wg.Wait()
}

func processMark(tag, msg string) func() {
	logger.I(tag, "[start],%v", msg)
	return func() {
		logger.I(tag, "[end],%v", msg)
	}
}

func (s *Server) dealExitSignal() {
	sg := make(chan os.Signal, 2)
	signal.Notify(sg, os.Interrupt, syscall.SIGTERM)
	<-sg
	s.Stop()
}

func (s *Server) Stop() {
	logger.I("Stop", "正在退出...")
	s.RunAfterServerStopFunc()
	time.Sleep(1 * time.Second)
	s.exit <- struct{}{}
}

func recoverProc() {
	if rec := recover(); rec != nil {
		if err, ok := rec.(error); ok {
			logger.E("PanicRecover", "Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
		} else {
			logger.E("PanicRecover", "Panic: %v\n stack:%v", rec, cast.ToString(debug.Stack()))
		}
		time.Sleep(1 * time.Second)
	}
}
