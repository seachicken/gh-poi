package shared

type Worktree struct {
	Path     string
	Branch   string
	IsMain   bool
	IsLocked bool
}
