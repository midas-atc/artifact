package midascli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var DEFAULT_CLUSTERCONFIG_PATH = "/mnt/data/.clusterconfig"
var CityNetAPI = "http://example.com:9002/datasets"

type midasCli struct {
	userConfig    *UserConfig
	clusterConfig *ClusterConfig
	// globalenv     *MIDASGlobalEnv
	prefix string
}

type Dataset struct {
	Name        string   `json: "name"`
	ID          string   `json: "_id"`
	CreateTime  string   `json: "create_time"`
	Files       []string `json: "files"`
	Labels      []string `json: "labels"`
	Description string   `json: "description"`
	Categories  string   `json: "categories"`
	Path        string   `json: "path"`
}

func (midascli *midasCli) UserConfig(option string) []string {
	var s []string
	switch strings.ToLower(option) {
	case "username":
		return append(s, midascli.userConfig.UserName)
	case "authfile":
		return append(s, midascli.userConfig.AuthFile)
	case "sshpath":
		return midascli.userConfig.SSHpath
	case "port":
		return append(s, midascli.userConfig.Port)
	case "path":
		return append(s, midascli.userConfig.path)
	default:
		log.Println("No option found in userconfig.")
		return s
	}
}
func (midascli *midasCli) ClusterConfig(option string) []string {
	var s []string
	switch strings.ToLower(option) {
	case "midasversion":
		return append(s, midascli.clusterConfig.midasVersion)
	case "dirs":
		return append(s, midascli.clusterConfig.Dirs["workdir"], midascli.clusterConfig.Dirs["userdir"])
	case "homedir":
		return append(s, midascli.clusterConfig.HomeDir)
	case "datasetdir":
		return append(s, midascli.clusterConfig.DatasetDir)
	case "conda":
		return append(s, midascli.clusterConfig.Conda)
	default:
		log.Println("No option found in clusterconfig.")
		return s
	}
}

func (midascli *midasCli) NewSession() *ssh.Session {
	buffer, err := ioutil.ReadFile(midascli.userConfig.AuthFile)
	if err != nil {
		log.Println("Failed to read AuthFile at ", midascli.userConfig.AuthFile)
		return nil
	}
	signer, _ := ssh.ParsePrivateKey(buffer)
	clientConfig := &ssh.ClientConfig{
		User: midascli.userConfig.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", midascli.userConfig.SSHpath[0], midascli.userConfig.Port), clientConfig)
	if err != nil {
		log.Println("Failed to dial: " + err.Error())
		return nil
	}
	session, err := client.NewSession()
	if err != nil {
		log.Println("Failed to create session: " + err.Error())
		return nil
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		log.Println("Failed to request for pseudo terminal: " + err.Error())
		return nil
	}
	return session
}

func (midascli *midasCli) NewPrefix() {
	if len(midascli.userConfig.SSHpath) < 2 {
		midascli.prefix = ""
	} else {
		var str string
		for _, s := range midascli.userConfig.SSHpath[1:] {
			str = str + fmt.Sprintf("ssh -A -t %s@%s ", midascli.userConfig.UserName, s)
		}
		midascli.prefix = str
	}
}

func NewmidasCli(userConfig *UserConfig, clusterConfig *ClusterConfig) *midasCli {
	midascli := &midasCli{
		userConfig:    userConfig,
		clusterConfig: clusterConfig,
	}
	midascli.NewPrefix()
	return midascli
}

func (midascli *midasCli) RemoteExecCmd(cmd string) bool {
	sess := midascli.NewSession()
	if sess == nil {
		log.Println("Failed to create remote session")
		os.Exit(-1)
	}
	w, err := sess.StdinPipe()
	if err != nil {
		log.Println("Failed to create StdinPipe", err)
		return true
	}
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr

	if err := sess.Run(cmd); err != nil {
		log.Println("Failed to run cmd \"", cmd, "\"", err)
		w.Close()
		return true
	}
	defer sess.Close()

	errors := make(chan error)
	go func() {
		errors <- sess.Wait()
	}()
	fmt.Fprint(w, "\x00")
	w.Close()
	return false
}

func (midascli *midasCli) RemoteExecCmdOutput(cmd string) ([]byte, bool) {
	sess := midascli.NewSession()
	if sess == nil {
		log.Println("Failed to create remote session")
		os.Exit(-1)
	}
	w, err := sess.StdinPipe()
	if err != nil {
		log.Println("Failed to create StdinPipe", err)
		return nil, true
	}
	var b bytes.Buffer

	sess.Stdout = &b
	sess.Stderr = os.Stderr

	if err := sess.Run(cmd); err != nil {
		log.Println("Failed to run cmd \"", cmd, "\"", err)
		w.Close()
		return nil, true
	}
	defer sess.Close()

	errors := make(chan error)
	go func() {
		errors <- sess.Wait()
	}()
	fmt.Fprint(w, "\x00")
	w.Close()
	return b.Bytes(), false
}

func (midascli *midasCli) UploadToUserDir(iscover bool, src string, dstDir string) (string, bool) {
	_, err := os.Stat(src)
	if err != nil {
		log.Printf("Failed to send to cluster. %s not exists.\n", src)
		return "", true
	}

	var cmd *exec.Cmd
	dst := midascli.userConfig.SSHpath[0]
	dst = fmt.Sprintf("%s@%s:%s", midascli.userConfig.UserName, dst, filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"], dstDir))
	ssh_config := fmt.Sprintf("/usr/bin/ssh -p %s -i %s", midascli.userConfig.Port, midascli.userConfig.AuthFile)
	if iscover {
		cmd = exec.Command("rsync", "-av", "--progress", "--delete", "--rsh", ssh_config, src, dst)
	} else {
		cmd = exec.Command("rsync", "-av", "--progress", "--rsh", ssh_config, src, dst)
	}

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Println("Failed to bind StdoutPipe")
		return dst, true
	}

	if err = cmd.Start(); err != nil {
		log.Println("Failed to start command")
		return dst, true
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return dst, true
	}
	return dst, false
}

func (midascli *midasCli) UploadToWorkerDir(dirName string, src string) (string, bool) {
	_, err := os.Stat(src)
	if err != nil {
		log.Printf("Failed to send to cluster. %s not exists.\n", src)
		return "", true
	}

	dst := midascli.userConfig.SSHpath[0]
	dst = fmt.Sprintf("%s@%s:%s", midascli.userConfig.UserName, dst, filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["workdir"]))
	ssh_config := fmt.Sprintf("/usr/bin/ssh -p %s -i %s", midascli.userConfig.Port, midascli.userConfig.AuthFile)
	cmd := exec.Command("rsync", "-av", "--progress", "--delete", "--rsh", ssh_config, src, dst)

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return dst, true
	}

	if err = cmd.Start(); err != nil {
		return dst, true
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return dst, true
	}
	return dst, false
}

// SCP from SSHPath[0] to localhost
func (midascli *midasCli) RecvFromCluster(src string, dst string, IsDir bool) bool {
	srcIP := midascli.userConfig.SSHpath[0]
	srcPath := fmt.Sprintf("%s@%s:%s", midascli.userConfig.UserName, srcIP, src)
	dstPath := fmt.Sprintf("%s", dst)

	var cmd *exec.Cmd
	if IsDir {
		cmd = exec.Command("scp", "-P", midascli.userConfig.Port, "-r", "-i", midascli.userConfig.AuthFile, srcPath, dstPath)
	} else {
		cmd = exec.Command("scp", "-P", midascli.userConfig.Port, "-i", midascli.userConfig.AuthFile, srcPath, dstPath)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println("Failed to run cmd in RecvFromCluster ", err.Error(), stderr.String())
		return true
	}
	return false
}

func (midascli *midasCli) BuildEnv(submitEnv *MIDASGlobalEnv, args ...string) map[string]string {
	var config MidasJobConfig
	localWorkDir, repoName, MIDASDir, datasets, err := config.ParseMidasJobConf(midascli, submitEnv, args)
	randString := RandString(16)
	if err == true {
		log.Println("Failed to parse MidasJob.conf")
		os.Exit(-1)
	}
	var envName string
	if config.Environment.Name == "" {
		hashString := config.EnvNameGenerator()
		// fmt.Fprintln(w, fmt.Sprintf("name: %s", config.Environment.Name + "-" + hashString))
		envName = hashString
	} else {
		envName = config.Environment.Name
	}

	if err = midascli.UploadRepo(repoName, localWorkDir); err == true {
		log.Println("Failed to upload env repository")
		os.Exit(-1)
	}

	if err = midascli.AddSoftLink(datasets); err == true {
		log.Println("Failed to add softlink")
		os.Exit(-1)
	}

	// Remove auto-generated files
	if err = midascli.RemoveAutoFiles(submitEnv); err == true {
		log.Println("Failed to remove all auto-generated files")
		os.Exit(-1)
	}

	// Generate env name and check if hit the cache, if so, return, otherwise, create new env.
	if midascli.CondaCacheCheck(envName) {
		homeDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName)
		condaBin := filepath.Join(homeDir, midascli.clusterConfig.Conda)
		condaYaml := filepath.Join(homeDir, midascli.clusterConfig.Dirs["workdir"], repoName, "configurations", "conda.yaml")
		cmd := fmt.Sprintf("%s %s env update -f %s -n %s\n", midascli.prefix, condaBin, condaYaml, envName)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to update %s\n", envName)
			os.Exit(-1)
			return MIDASDir
		}
		fmt.Printf("Env %s exists, dependencies updated.\n", envName)
		return MIDASDir
	}
	if err = midascli.CondaCreate(repoName, envName, randString); err == true {
		log.Println("Failed to create conda env")
		os.Exit(-1)
	}
	return MIDASDir
}

func (midascli *midasCli) UploadRepo(repoName string, localWorkDir string) bool {
	dst, err := midascli.UploadToWorkerDir(repoName, localWorkDir)
	if err == true {
		log.Println("Failed to upload repo to ", dst)
		return true
	}
	// fmt.Println("Successfully upload repo to ", dst)
	return false
}

func (midascli *midasCli) AddSoftLink(datasets []string) bool {
	for _, s := range datasets {
		cmd := fmt.Sprintf("curl -X GET %s/%s", CityNetAPI, s)

		out, err := midascli.RemoteExecCmdOutput(cmd)
		if err == true {
			log.Println("Failed to access CityNet API")
			return true
		}

		var config Dataset
		json.Unmarshal(out, &config)

		datasetpath := filepath.Join(midascli.clusterConfig.DatasetDir, config.Path)
		remoteUserDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
		remoteDir := filepath.Join(remoteUserDir, config.Name)

		cmd = fmt.Sprintf("%s rm -f %s", midascli.prefix, remoteDir)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to remove old softlink at %s\n", remoteDir)
			return true
		}
		cmd = fmt.Sprintf("%s ln -s %s %s", midascli.prefix, datasetpath, remoteDir)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Println("Failed to add softlink to user directory", err)
			return true
		}

		// fmt.Println("Successfully create softlink at", remoteDir)
	}
	return false
}

func (midascli *midasCli) RemoveAutoFiles(submitEnv *MIDASGlobalEnv) bool {
	if err := os.RemoveAll(submitEnv.LocalConfDir); err != nil {
		log.Println("Failed to remove auto-generated files")
		return true
	}
	if err := os.Remove(filepath.Join(submitEnv.LocalWorkDir, "run.sh")); err != nil {
		log.Println("Failed to remove auto-generated run.sh")
		return true
	}
	return false
}

func (midascli *midasCli) CondaCreate(repoName string, envName string, randString string) bool {
	homeDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName)
	condaBin := filepath.Join(homeDir, midascli.clusterConfig.Conda)
	condaYaml := filepath.Join(homeDir, midascli.clusterConfig.Dirs["workdir"], repoName, "configurations", "conda.yaml")
	cmd := fmt.Sprintf("%s %s env create -f %s -n %s\n", midascli.prefix, condaBin, condaYaml, envName)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to run cmd in CondaCreate")
		return true
	}

	fmt.Printf("Successfully create environment %s\n.", envName)
	return false
}

func (midascli *midasCli) CondaRemove(envName string) bool {
	homeDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName)
	condaBin := filepath.Join(homeDir, midascli.clusterConfig.Conda)
	cmd := fmt.Sprintf("%s %s remove -n %s --all -y", midascli.prefix, condaBin, envName)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to run cmd in CondaRemove")
		return true
	}
	// fmt.Println("Previous environment \"", envName, "\" removed.")
	return false
}

func RandString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6
		letterIdxMask = 1<<letterIdxBits - 1
		letterIdxMax  = 63 / letterIdxBits
	)
	var src = rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

func (midascli *midasCli) XSubmit(args ...string) bool {
	var submitEnv = NewGlobalEnv()

	cmd := fmt.Sprintf("mkdir -p  %s", filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["workdir"]))
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to create remote workdir")
		return true
	}
	cmd = fmt.Sprintf("mkdir -p  %s", filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"]))
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to create remote userdir")
		return true
	}
	cmd = fmt.Sprintf("mkdir -p  %s", filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"], submitEnv.SlurmUserlog))
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to create remote workdir")
		return true
	}

	MIDASDir := midascli.BuildEnv(submitEnv, args...)

	cmd = fmt.Sprintf("%s sbatch %s", midascli.prefix, filepath.Join(submitEnv.RemoteWorkDir, "configurations", "run.slurm"))

	// Create `RUNDIR` in remote and run cmd at `RUNDIR`
	cmd = fmt.Sprintf("mkdir -p %s && cd %s && %s", MIDASDir["MIDAS_WORKDIR"], MIDASDir["MIDAS_WORKDIR"], cmd)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to run cmd in midas submit")
		return true
	}
	fmt.Println("Job", submitEnv.RepoName, "submitted.")
	return false
}

func (midascli *midasCli) XPS(job string, args ...string) bool {
	var cmd string
	if job == "" {
		cmd = fmt.Sprintf("%s squeue", midascli.prefix)
	} else {
		cmd = fmt.Sprintf("%s squeue -j %s", midascli.prefix, job)
	}
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to run cmd in midas ps")
		return true
	}
	return false
}

func (midascli *midasCli) XInit(args ...string) bool {
	// Remote receive config file
	src := DEFAULT_CLUSTERCONFIG_PATH
	dst := fmt.Sprintf("%s", filepath.Join(os.Getenv("HOME"), ".midas"))
	IsDir := false

	if err := midascli.RecvFromCluster(src, dst, IsDir); err == true {
		log.Println("Failed to receive config files from MIDAS")
		return true
	}
	cmd := fmt.Sprintf("sinfo")
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to get cluster information")
		return true
	}
	return false
}

func (midascli *midasCli) XAdd(args ...string) bool {
	// Add new dependency to MidasJob.conf
	var config MidasJobConfig
	err := config.AddDepMidasJobFile(midascli, args)
	if err == true {
		log.Println("Failed to add dependency to MidasJob.conf")
		os.Exit(-1)
	}
	return false
}

func (midascli *midasCli) XInstall(args ...string) bool {
	var config MidasJobConfig
	var submitEnv *MIDASGlobalEnv
	_, _, _, _, err := config.ParseMidasJobConf(midascli, submitEnv, args)
	if err == true {
		log.Println("Failed to parse MidasJob.conf")
		os.Exit(-1)
	}
	condaYaml := filepath.Join(".", "configurations", "conda.yaml")
	removeCmd := exec.Command(midascli.clusterConfig.Conda, "env", "remove", "-n", config.Environment.Name)
	if out, err := removeCmd.CombinedOutput(); err != nil {
		log.Println("Failed to create local environment. Err: ", err.Error())
		return true
	} else {
		fmt.Printf("%s\n", string(out))
	}
	createCmd := exec.Command(midascli.clusterConfig.Conda, "env", "create", "-f", condaYaml)
	if out, err := createCmd.CombinedOutput(); err != nil {
		log.Println("Failed to create local environment. Err: ", err)
		return true
	} else {
		fmt.Printf("%s\n", string(out))
	}
	fmt.Println("Successfully create environment locally.")
	return false
}

func (midascli *midasCli) XUpload(iscover bool, args ...string) bool {
	var src, dst string
	src = args[0]
	if len(args) > 1 {
		dst = args[1]
	} else {
		dst = "."
	}

	if _, err := midascli.UploadToUserDir(iscover, src, dst); err {
		log.Printf("Failed to upload %s to %s\n", src, dst)
		return true
	}
	return false
}

func (midascli *midasCli) XDownload(IsDir bool, args ...string) bool {
	var src, dst, remotesrc string
	src = args[0]
	if len(args) > 1 {
		dst = args[1]
	} else {
		dst = "."
	}

	// Format src, dst
	if src[0:1] == "./" {
		remotesrc = src[2:]
	} else if src[0] == '/' {
		remotesrc = src[1:]
	} else {
		remotesrc = src
	}

	remoteUserDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	remotesrc = filepath.Join(remoteUserDir, remotesrc)

	if err := midascli.RecvFromCluster(remotesrc, dst, IsDir); err {
		if IsDir {
			log.Printf("Failed to receive directory %s to %s\n", src, dst)
			return true
		} else {
			log.Printf("Failed to receive file %s to %s\n", src, dst)
			return true
		}
	}
	return false
}

// Only allow remote workdir copy to remote userdir
// Src must contain repoName first
func (midascli *midasCli) XCP(IsDir bool, args ...string) bool {
	var src, dst, remotesrc, remotedst string
	src = args[0]
	dst = args[1]

	// Format src, dst
	if src[0:1] == "./" {
		remotesrc = src[2:]
	} else if src[0] == '/' {
		remotesrc = src[1:]
	} else {
		remotesrc = src
	}

	if dst[0:1] == "./" {
		remotedst = dst[2:]
	} else if dst[0] == '/' {
		remotedst = dst[1:]
	} else {
		remotedst = dst
	}

	remoteWorkDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["workdir"])
	remoteUserDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	remotesrc = filepath.Join(remoteWorkDir, remotesrc)
	remotedst = filepath.Join(remoteUserDir, remotedst)

	if IsDir {
		cmd := fmt.Sprintf("cp -r %s %s", remotesrc, remotedst)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to copy %s to %s\n", src, dst)
			return true
		}
	} else {
		cmd := fmt.Sprintf("mkdir -p %s && cp %s %s", remotedst, remotesrc, remotedst)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to copy %s to %s\n", src, dst)
			return true
		}
	}
	return false
}

func (midascli *midasCli) XLS(IsLong bool, IsReverse bool, IsAll bool, args ...string) bool {
	var src, flags string
	if len(args) > 0 {
		src = args[0]
	} else {
		src = "."
	}
	flags = ""
	if IsLong {
		flags += " -l"
	}
	if IsReverse {
		flags += " -r"
	}
	if IsAll {
		flags += " -a"
	}

	remoteUserDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	remote := filepath.Join(remoteUserDir, src)

	cmd := fmt.Sprintf("ls %s %s", flags, remote)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Printf("Failed to ls %s %s\n", flags, remote)
		return true
	}
	return false
}

func (midascli *midasCli) XENVLS(IsEnv bool, args ...string) bool {
	homeDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName)
	condaBin := filepath.Join(homeDir, midascli.clusterConfig.Conda)
	var src, flags string
	if len(args) > 0 {
		src = args[0]
	} else {
		src = ""
	}
	flags = ""
	if IsEnv {
		flags += "-n"
		cmd := fmt.Sprintf("%s list %s %s", condaBin, flags, src)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to list environment packages%s %s\n", flags, src)
			return true
		}
	} else {
		cmd := fmt.Sprintf("%s env list %s", condaBin, src)
		if err := midascli.RemoteExecCmd(cmd); err == true {
			log.Printf("Failed to list environment %s %s\n", flags, src)
			return true
		}
	}
	return false
}

func (midascli *midasCli) XCancel(job string, args ...string) bool {
	var cmd string
	cmd = fmt.Sprintf("%s scancel %s", midascli.prefix, job)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to cancel job ", job)
		return true
	}
	return false
}

func (midascli *midasCli) XDataset(args ...string) bool {
	if err := midascli.AddSoftLink(args); err == true {
		log.Println("Failed to create dataset ", args[0])
		return true
	}
	return false
}

func (midascli *midasCli) XCat(args ...string) bool {
	var cmd string
	remoteUserDir := filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, midascli.clusterConfig.Dirs["userdir"])
	remote := filepath.Join(remoteUserDir, args[0])

	cmd = fmt.Sprintf("cat %s", remote)
	if err := midascli.RemoteExecCmd(cmd); err == true {
		log.Println("Failed to cat file ", args[0])
		return true
	}
	return false
}

func (midascli *midasCli) CondaCacheCheck(envName string) bool {
	// Get env list from remote
	cmd := fmt.Sprintf("ls -ltr %s", filepath.Join(midascli.clusterConfig.HomeDir, midascli.userConfig.UserName, ".Miniconda3", "envs"))
	var envList []string
	if out, err := midascli.RemoteExecCmdOutput(cmd); err == true {
		log.Println("Failed to get env list")
		os.Exit(1)
	} else {
		envList = strings.Split(strings.Trim(string(out), "\n"), "\n")
		for i, env := range envList {
			if i > 0 {
				splitString := strings.Split(env, " ")
				envList[i] = strings.Trim(splitString[len(splitString)-1], string(13))
			}
		}
		envList = envList[1:]
	}
	// Check if there is a hit, if so, return true, otherwise, return false
	for _, env := range envList {
		if env == envName {
			return true
		}
	}
	// Check the env cach length, if length > 10, remove the older env.
	envList = append(envList, envName)
	for {
		if len(envList) <= 10 {
			break
		}
		if err := midascli.CondaRemove(envList[0]); err == true {
			log.Println("Failed to remove conda env")
			os.Exit(-1)
		}
		envList = envList[1:]
	}
	return false
}

func (midascli *midasCli) XTest(args ...string) bool {
	var config MidasJobConfig
	var src string
	src = args[0]

	err := config.DirSizeCheck(src, midascli)
	log.Println("Error:", err)
	log.Println(midascli.clusterConfig.StorageQuota)
	return false
}
