package shared

// NoRepoMsg is the user-facing message printed when a command is run outside
// of a git repository.
const NoRepoMsg = "must be run from inside a git repository"

// ProtectDeprecationMsg is printed when the deprecated `protect` command is used.
const ProtectDeprecationMsg = "warning: 'protect' is deprecated, please use 'lock' instead"

// UnprotectDeprecationMsg is printed when the deprecated `unprotect` command is used.
const UnprotectDeprecationMsg = "warning: 'unprotect' is deprecated, please use 'unlock' instead"
