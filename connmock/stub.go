package connmock

import (
	"io/ioutil"
	"path"
	"runtime"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/mocks"
)

type (
	Stub struct {
		Conn *mocks.MockConnection
		t    gomock.TestHelper
	}

	Times struct {
		N int
	}

	Conf struct {
		Times *Times
	}

	LogStub struct {
		BranchName string
		Result     string
	}
)

var (
	fixturePath = "fixtures"
)

func Setup(ctrl *gomock.Controller) *Stub {
	conn := mocks.NewMockConnection(ctrl)
	return &Stub{conn, ctrl.T}
}

func NewConf(times *Times) *Conf {
	return &Conf{
		times,
	}
}

func (s *Stub) CheckRepos(err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			CheckRepos(gomock.Any(), gomock.Any()).
			Return(err),
		conf,
	)
	return s
}

func (s *Stub) GetRepoNames(path string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetRepoNames().
			Return(s.readFile("gh", "repo", path), err),
		conf,
	)
	return s
}

func (s *Stub) GetBranchNames(name string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.EXPECT().
			GetBranchNames().
			Return(s.readFile("git", "branch", name), err),
		conf,
	)
	return s
}

func (s *Stub) GetLog(stubs []LogStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	for _, stub := range stubs {
		configure(
			s.Conn.EXPECT().
				GetLog(stub.BranchName).
				Return(s.readFile("git", "log", stub.Result), err),
			conf,
		)
	}
	return s
}

func (s *Stub) GetPullRequests(path string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetPullRequests(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(s.readFile("gh", "pr", path), err),
		conf,
	)
	return s
}

func (s *Stub) DeleteBranches(err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			DeleteBranches(gomock.Any()).
			Return("", err),
		conf,
	)
	return s
}

func configure(call *gomock.Call, conf *Conf) {
	if conf == nil {
		return
	}

	if conf.Times == nil {
		call.AnyTimes()
	} else {
		call.Times(conf.Times.N)
	}
}

func (s *Stub) readFile(command string, category string, name string) string {
	_, filename, _, _ := runtime.Caller(0)

	ext := ".txt"
	if command == "gh" {
		ext = ".json"
	}
	b, err := ioutil.ReadFile(path.Join(filename, "../..", fixturePath, command, category+"_"+name+ext))
	if err != nil {
		s.t.Fatalf("%v", err)
	}
	return string(b)
}
