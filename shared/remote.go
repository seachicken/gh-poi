package shared

type Remote struct {
	Name       string
	Hostname   string
	RepoName   string
	GhResolved string
}

func (r Remote) ResolvedRepoName() string {
	if len(r.GhResolved) > 1 && r.GhResolved != "base" {
		return r.GhResolved
	}
	return r.RepoName
}
