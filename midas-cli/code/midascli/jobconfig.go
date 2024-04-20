package midascli

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type MidasJobConfig struct {
	Entrypoint  []string
	Environment struct {
		Name         string
		Channels     []string
		Dependencies []string
	}
	Job struct {
		Name    string
		General []string
		Module  []string
		Env     []string
		Log     string
		// Model 	string
	}
	Datasets []string
}

var CONDA_SHELL_PATH = ".Miniconda3/etc/profile.d/conda.sh"
var MidasJobFile = "MidasJob.conf"
var confDir = "configurations"
var RunshFile = "run.sh"

func (config *MidasJobConfig) MIDASJobEnv(remoteWorkDir string, remoteUserDir string) ([]string, map[string]string) {
	var strlist []string
	MIDASDir := make(map[string]string)
	// MIDAS Global Env
	strlist = append(strlist, fmt.Sprintf("MIDAS_WORKDIR=%s", remoteWorkDir))
	MIDASDir["MIDAS_WORKDIR"] = remoteWorkDir
	strlist = append(strlist, fmt.Sprintf("MIDAS_USERDIR=%s", remoteUserDir))
	MIDASDir["MIDAS_USERDIR"] = remoteUserDir
	return strlist, MIDASDir
}

func (config *MidasJobConfig) ParseMidasJobConf(midascli *midasCli, submitEnv *MIDASGlobalEnv, args []string) (string, string, map[string]string, []string, bool) {
	MIDASDir := make(map[string]string)
	if len(args) < 1 {
		submitEnv.LocalWorkDir, _ = filepath.Abs(path.Dir("."))
		if err := config.DirSizeCheck(submitEnv.LocalWorkDir, midascli); err == true {
			os.Exit(-1)
		}
		fmt.Println("Start parsing MidasJob.conf...")
		submitEnv.LocalConfDir = filepath.Join(submitEnv.LocalWorkDir, confDir)
		dirlist := strings.Split(submitEnv.LocalWorkDir, "/")
		submitEnv.RepoName = dirlist[len(dirlist)-1]
		submitEnv.RemoteWorkDir = filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["workdir"], submitEnv.RepoName)
		submitEnv.RemoteUserDir = filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	} else {
		submitEnv.LocalWorkDir, _ = filepath.Abs(args[0])
		if err := config.DirSizeCheck(submitEnv.LocalWorkDir, midascli); err == true {
			os.Exit(-1)
		}
		fmt.Println("Start parsing MidasJob.conf...")
		submitEnv.LocalConfDir = filepath.Join(submitEnv.LocalWorkDir, confDir)
		dirlist := strings.Split(submitEnv.LocalWorkDir, "/")
		submitEnv.RepoName = dirlist[len(dirlist)-1]
		submitEnv.RemoteWorkDir = filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["workdir"], submitEnv.RepoName)
		submitEnv.RemoteUserDir = filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	}
	LocalMidasJobFile := filepath.Join(submitEnv.LocalWorkDir, MidasJobFile)
	yamlFile, err := ioutil.ReadFile(LocalMidasJobFile)
	if err != nil {
		return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, nil, true
	}

	err = yaml.Unmarshal(yamlFile, config)
	if _, err = os.Stat(submitEnv.LocalConfDir); os.IsNotExist(err) {
		os.Mkdir(submitEnv.LocalConfDir, 0755)
	}

	if err := config.CondaFile(submitEnv); err == true {
		log.Println("Failed to generate Environment config file")
		return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, nil, true
	}
	var err1 bool
	if MIDASDir, err1 = config.SlurmFile(submitEnv); err1 == true {
		log.Println("Failed to generate Slurm config file")
		return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, config.Datasets, true
	}
	if err := config.CityFile(submitEnv); err == true {
		log.Println("Failed to generate Datasets config file")
		return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, config.Datasets, true
	}
	if err := config.RunshFile(midascli, submitEnv); err == true {
		log.Println("Failed to generate Run.sh exec file")
		return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, config.Datasets, true
	}
	return submitEnv.LocalWorkDir, submitEnv.RepoName, MIDASDir, config.Datasets, false
}

func (config *MidasJobConfig) CondaFile(submitEnv *MIDASGlobalEnv) bool {
	localConfDir := submitEnv.LocalConfDir
	f, err := os.Create(filepath.Join(localConfDir, "conda.yaml"))
	if err != nil {
		log.Println("Failed to create Conda config file")
		return true
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	// Conda file
	var EnvName string
	if config.Environment.Name == "" {
		hashString := config.EnvNameGenerator()
		// fmt.Fprintln(w, fmt.Sprintf("name: %s", config.Environment.Name + "-" + hashString))
		EnvName = hashString
	} else {
		EnvName = config.Environment.Name
	}
	fmt.Fprintln(w, fmt.Sprintf("name: %s", EnvName))
	// Channels
	fmt.Fprintln(w, fmt.Sprintf("channels:"))
	for _, s := range config.Environment.Channels {
		str := fmt.Sprintf("  - %s", s)
		fmt.Fprintln(w, str)
	}
	// Dependencies
	fmt.Fprintln(w, "dependencies:")
	for _, s := range config.Environment.Dependencies {
		str := fmt.Sprintf("  - %s", s)
		fmt.Fprintln(w, str)
	}
	w.Flush()
	return false
}

func (config *MidasJobConfig) SlurmFile(submitEnv *MIDASGlobalEnv) (map[string]string, bool) {
	localConfDir := submitEnv.LocalConfDir
	remoteWorkDir := submitEnv.RemoteWorkDir
	remoteUserDir := submitEnv.RemoteUserDir
	MIDASDir := make(map[string]string)
	f, err := os.Create(filepath.Join(localConfDir, "run.slurm"))
	if err != nil {
		log.Println("Failed to create Slurm config file")
		log.Fatal(err)
		return MIDASDir, true
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// Slurm file
	fmt.Fprintln(w, "#!/bin/bash")
	// SBATCH
	var CHECK_OUTPUT = false
	for _, s := range config.Job.General {
		if strings.Contains(s, "output=") == true {
			CHECK_OUTPUT = true
		}
		str := fmt.Sprintf("#SBATCH --%s", s)
		str = ReplaceGlobalEnv(str, submitEnv)
		fmt.Fprintln(w, str)
	}
	if CHECK_OUTPUT == false {
		str := fmt.Sprintf("#SBATCH --output=%s", "${MIDAS_SLURM_USERLOG}/slurm-%j.out")
		str = ReplaceGlobalEnv(str, submitEnv)
		fmt.Fprintln(w, str)
	}
	// Module
	for _, s := range config.Job.Module {
		str := fmt.Sprintf("module load %s", s)
		fmt.Fprintln(w, str)
	}
	// Env
	for _, s := range config.Job.Env {
		str := fmt.Sprintf("export %s", s)
		fmt.Fprintln(w, str)
	}

	// MIDAS Env
	strlist, MIDASDir := config.MIDASJobEnv(remoteWorkDir, remoteUserDir)
	for _, s := range strlist {
		str := fmt.Sprintf("export %s", s)
		fmt.Fprintln(w, str)
	}
	str := fmt.Sprintf("srun %s", filepath.Join(remoteWorkDir, RunshFile))
	fmt.Fprintln(w, str)
	w.Flush()
	return MIDASDir, false
}

func (config *MidasJobConfig) CityFile(submitEnv *MIDASGlobalEnv) bool {
	localConfDir := submitEnv.LocalConfDir
	f, err := os.Create(filepath.Join(localConfDir, "citynet.sh"))
	if err != nil {
		log.Println("Failed to create Datasets config file")
		return true
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, s := range config.Datasets {
		fmt.Fprintln(w, s)
	}

	w.Flush()
	return false
}

func (config *MidasJobConfig) RunshFile(midascli *midasCli, submitEnv *MIDASGlobalEnv) bool {
	localWorkDir := submitEnv.LocalWorkDir
	localRunshFile := filepath.Join(localWorkDir, RunshFile)
	f, err := os.Create(localRunshFile)
	if err != nil {
		log.Println("Failed to create run.sh file")
		return true
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	homeDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName)
	str := fmt.Sprintf("#!/bin/bash\nsource %s/%s", homeDir, CONDA_SHELL_PATH)
	fmt.Fprintln(w, str)

	var EnvName string
	if config.Environment.Name == "" {
		hashString := config.EnvNameGenerator()
		// fmt.Fprintln(w, fmt.Sprintf("name: %s", config.Environment.Name + "-" + hashString))
		EnvName = hashString
	} else {
		EnvName = config.Environment.Name
	}
	str = fmt.Sprintf("conda activate %s\n", EnvName)
	fmt.Fprintln(w, str)

	for _, s := range config.Entrypoint {
		str = fmt.Sprintf("%s \\", s)
		fmt.Fprintln(w, str)
	}
	w.Flush()
	if err = os.Chmod(localRunshFile, 0755); err != nil {
		log.Println("Failed to chmod run.sh")
		return true
	}
	return false
}
func (config *MidasJobConfig) AddDepMidasJobFile(midascli *midasCli, args []string) bool {
	var MidasJobFile string
	MidasJobFile = "MidasJob.conf"
	yamlFile, err := ioutil.ReadFile(MidasJobFile)
	if err != nil {
		log.Println("Failed to read file")
		return true
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Println("Failed to parse original yaml file")
		return true
	}
	for i := 0; i < len(config.Environment.Dependencies); i++ {
		slist := strings.Split(config.Environment.Dependencies[i], "=")
		deplist := strings.Split(args[0], "=")
		if deplist[0] == slist[0] {
			fmt.Println("Remove the original dependency", config.Environment.Dependencies[i])
			config.Environment.Dependencies = append(config.Environment.Dependencies[:i], config.Environment.Dependencies[i+1:]...)
		}
	}
	config.Environment.Dependencies = append(config.Environment.Dependencies, args[0])
	yamlFile, err = yaml.Marshal(config)
	if err != nil {
		log.Println("Failed to format file")
		return true
	}
	err = ioutil.WriteFile(MidasJobFile, yamlFile, 0755)
	if err != nil {
		log.Println("Failed to write file")
		return true
	}

	return false
}

// Currently only replace MIDAS_WORKDIR, MIDAS_USERDIR
func ReplaceGlobalEnv(str string, submitEnv *MIDASGlobalEnv) string {
	str = strings.Replace(str, "${MIDAS_WORKDIR}", submitEnv.RemoteWorkDir, -1)
	str = strings.Replace(str, "$MIDAS_WORKDIR", submitEnv.RemoteWorkDir, -1)
	str = strings.Replace(str, "${MIDAS_USERDIR}", submitEnv.RemoteUserDir, -1)
	str = strings.Replace(str, "$MIDAS_USERDIR", submitEnv.RemoteUserDir, -1)

	slurm_log := filepath.Join(submitEnv.RemoteUserDir, submitEnv.SlurmUserlog)
	str = strings.Replace(str, "${MIDAS_SLURM_USERLOG}", slurm_log, -1)
	str = strings.Replace(str, "$MIDAS_SLURM_USERLOG", slurm_log, -1)
	return str
}
func (config *MidasJobConfig) EnvNameGenerator() string {
	// Parse package (with version) list from conda.yaml
	dep := config.Environment.Dependencies
	// Sort the package by Alphabetical order and contact as a string
	sort.Strings(dep)
	// Generate md5 hash value from the string
	jointDep := strings.Join(dep, " ")
	data := []byte(jointDep)
	hashValue := md5.Sum(data)
	hashString := hex.EncodeToString(hashValue[:])
	return hashString
}

func (config *MidasJobConfig) DirSizeCheck(path string, midascli *midasCli) bool {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	size = size / 1024 / 1024
	if err != nil {
		log.Printf("Failed to calculate folder size .\n")
		return true
	}
	if size > midascli.clusterConfig.StorageQuota {
		log.Printf("Fail to upload file. Upload size:%dMB. Limitation: %dMB.\n", size, midascli.clusterConfig.StorageQuota)
		return true
	}
	return false
}
