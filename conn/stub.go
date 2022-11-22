package conn

import (
	"os"
	"path/filepath"
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

	RemoteHeadStub struct {
		BranchName string
		Filename   string
	}

	LsRemoteHeadStub struct {
		BranchName string
		Filename   string
	}

	AssociatedBranchNamesStub struct {
		Oid      string
		Filename string
	}

	LogStub struct {
		BranchName string
		Filename   string
	}

	ConfigStub struct {
		BranchName string
		Filename   string
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
			CheckRepos(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(err),
		conf,
	)
	return s
}

func (s *Stub) GetRemoteNames(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetRemoteNames(gomock.Any()).
			Return(s.readFile("git", "remote", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetSshConfig(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetSshConfig(gomock.Any(), gomock.Any()).
			Return(s.readFile("ssh", "config", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetRepoNames(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetRepoNames(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(s.readFile("gh", "repo", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetBranchNames(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.EXPECT().
			GetBranchNames(gomock.Any()).
			Return(s.readFile("git", "branch", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetMergedBranchNames(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.EXPECT().
			GetMergedBranchNames(gomock.Any(), "origin", "main").
			Return(s.readFile("git", "branchMerged", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetRemoteHeadOid(stubs []RemoteHeadStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	if stubs == nil {
		configure(
			s.Conn.EXPECT().
				GetRemoteHeadOid(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", err),
			conf,
		)
	} else {
		for _, stub := range stubs {
			configure(
				s.Conn.EXPECT().
					GetRemoteHeadOid(gomock.Any(), gomock.Any(), stub.BranchName).
					Return(s.readFile("git", "remoteHead", stub.Filename), err),
				conf,
			)
		}
	}
	return s
}

func (s *Stub) GetLsRemoteHeadOid(stubs []LsRemoteHeadStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	if stubs == nil {
		configure(
			s.Conn.EXPECT().
				GetLsRemoteHeadOid(gomock.Any(), gomock.Any(), gomock.Any()).
				Return("", err),
			conf,
		)
	} else {
		for _, stub := range stubs {
			configure(
				s.Conn.EXPECT().
					GetLsRemoteHeadOid(gomock.Any(), gomock.Any(), stub.BranchName).
					Return(s.readFile("git", "lsRemoteHead", stub.Filename), err),
				conf,
			)
		}
	}
	return s
}

func (s *Stub) GetAssociatedRefNames(stubs []AssociatedBranchNamesStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	for _, stub := range stubs {
		configure(
			s.Conn.EXPECT().
				GetAssociatedRefNames(gomock.Any(), stub.Oid).
				Return(s.readFile("git", "abranch", stub.Filename), err),
			conf,
		)
	}
	return s
}

func (s *Stub) GetLog(stubs []LogStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	for _, stub := range stubs {
		configure(
			s.Conn.EXPECT().
				GetLog(gomock.Any(), stub.BranchName).
				Return(s.readFile("git", "log", stub.Filename), err),
			conf,
		)
	}
	return s
}

func (s *Stub) GetPullRequests(filename string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetPullRequests(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(s.readFile("gh", "pr", filename), err),
		conf,
	)
	return s
}

func (s *Stub) GetUncommittedChanges(uncommittedChanges string, err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			GetUncommittedChanges(gomock.Any()).
			Return(uncommittedChanges, err),
		conf,
	)
	return s
}

func (s *Stub) GetConfig(stubs []ConfigStub, err error, conf *Conf) *Stub {
	s.t.Helper()
	for _, stub := range stubs {
		configure(
			s.Conn.
				EXPECT().
				GetConfig(gomock.Any(), stub.BranchName).
				Return(s.readFile("git", "config", stub.Filename), err),
			conf,
		)
	}
	return s
}

func (s *Stub) CheckoutBranch(err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			CheckoutBranch(gomock.Any(), gomock.Any()).
			Return("", err),
		conf,
	)
	return s
}

func (s *Stub) DeleteBranches(err error, conf *Conf) *Stub {
	s.t.Helper()
	configure(
		s.Conn.
			EXPECT().
			DeleteBranches(gomock.Any(), gomock.Any()).
			Return("", err),
		conf,
	)
	return s
}

func configure(call *gomock.Call, conf *Conf) {
	if conf == nil || conf.Times == nil {
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
	b, err := os.ReadFile(filepath.Join(filename, "..", fixturePath, command, category+"_"+name+ext))
	if err != nil {
		s.t.Fatalf("%v", err)
	}
	return string(b)
}
