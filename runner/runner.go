// base code in this file is from https://github.com/mraron/njudge. give them a star if you like.
package runner

import (
	"bytes"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Runner interface {
	TimeLimit(duration time.Duration) Runner
	MemoryLimit(limit int) Runner
}

type Dummy struct {
	tmpdir string
	env    []string
	tl     time.Duration

	stdin          io.Reader
	stdout, stderr io.Writer

	workingDir string
}

func NewDummy() *Dummy {
	return &Dummy{}
}

func (s *Dummy) Id() string {
	return s.tmpdir
}

func (s *Dummy) Init() error {
	var err error
	if s.tmpdir, err = os.MkdirTemp("", "dummysandbox"); err != nil {
		return err
	}

	s.workingDir = s.tmpdir
	err = os.Chdir(s.tmpdir)
	return err
}

func (s *Dummy) Pwd() string {
	return s.tmpdir
}

func (s *Dummy) CreateFile(filename string, r io.Reader) error {
	f, err := os.Create(filename)
	if err != nil {
		pkg.Log.Debug("Error occurred while creating file ", err)
		return err
	}

	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}

	return f.Close()
}

func (s *Dummy) GetFile(name string) (io.Reader, error) {
	f, err := ioutil.ReadFile(filepath.Join(s.Pwd(), name))
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(f), nil
}

func (s *Dummy) MakeExecutable(filename string) error {

	err := os.Chmod(filename, 0777)

	return err
}

func (s *Dummy) SetMaxProcesses(i int) Runner {
	return s
}

func (s *Dummy) TimeLimit(tl time.Duration) Runner {
	s.tl = tl
	return s
}

func (s *Dummy) MemoryLimit(int) Runner {
	return s
}

func (s *Dummy) Stdin(reader io.Reader) Runner {
	s.stdin = reader
	return s
}

func (s *Dummy) Stderr(writer io.Writer) Runner {
	s.stderr = writer
	return s
}

func (s *Dummy) Stdout(writer io.Writer) Runner {
	s.stdout = writer
	return s
}

func (s *Dummy) Run(prg string, needStatus bool) (structs.Verdict, error) {
	cmd := exec.Command("python3", prg)
	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	cmd.Dir = s.workingDir
	//cmd.Env = append(s.env, "PATH="+os.Getenv("PATH")+":"+s.tmpdir)

	var (
		errKill, errWait error
		finish           = make(chan bool, 1)
		wg               sync.WaitGroup
	)

	start := time.NewTimer(s.tl)
	if err := cmd.Start(); err != nil {
		return structs.VerdictUnknown, err
	}
	defer start.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		errWait = cmd.Wait()
		finish <- true
	}()

	v := structs.VerdictOK
	select {
	case <-start.C:
		v = structs.VerdictTimeLimit
		if errKill = cmd.Process.Kill(); errKill != nil {
			v = structs.VerdictUnknown
		}
	case <-finish:
	}

	wg.Wait()

	//if errWait != nil && (strings.HasPrefix(errWait.Error(), "exit status") || strings.HasPrefix(errWait.Error(), "signal:")) {
	if _, ok := errWait.(*exec.ExitError); ok {
		if v == structs.VerdictOK {
			v = structs.VerdictRuntimeError
		}
		errWait = nil
	}

	if errWait != nil {
		return v, errWait
	}

	return v, errKill
}

func (s *Dummy) Cleanup() error {
	return os.RemoveAll(s.tmpdir)
}

func RunTask(timeLimit time.Duration, memoryLimit int, code string, input io.Reader, output io.Writer, stderr io.Writer) (structs.Verdict, error) {
	// runner for python
	d := NewDummy()
	err := d.Init()
	if err != nil {
		return structs.VerdictUnknown, err
	}

	d.MemoryLimit(memoryLimit)
	d.TimeLimit(timeLimit)
	d.Stdin(input)
	d.Stdout(output)
	d.Stderr(stderr)
	const fileName = "runner_test.py"
	err = d.CreateFile(fileName, strings.NewReader(code))
	if err != nil {
		return structs.VerdictUnknown, err
	}
	err = d.MakeExecutable(fileName)
	if err != nil {
		return structs.VerdictUnknown, err
	}

	v, err := d.Run("runner_test.py", false)
	if err != nil {
		return v, err
	}
	err = d.Cleanup()
	if err != nil {
		pkg.Log.Warning("error on doing cleanup of runner: ", err)
	}
	return v, nil
}
