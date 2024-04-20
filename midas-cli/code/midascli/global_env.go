package midascli

type MIDASGlobalEnv struct {
	RepoName      string
	LocalWorkDir  string
	LocalConfDir  string
	RemoteWorkDir string
	RemoteUserDir string

	SlurmUserlog string
}

var DEFAULT_SLURMDIR = "slurm_log"

func NewGlobalEnv() *MIDASGlobalEnv {
	var env MIDASGlobalEnv
	env.SlurmUserlog = DEFAULT_SLURMDIR
	return &env
}
