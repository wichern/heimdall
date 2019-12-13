package scriptrunner

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
)

/*
 * Changes:
 *   - Cmd can only be used once per call. Thus create a new one on every call.
 *   - Use channels to communicate the state of the run to the script struct.
 */

// Script contains all information of a runnable script.
type Script struct {
	ID      int    // Unique id of script.
	Name    string // Basename of executable.
	cmds    []*exec.Cmd
	path    string
	stdout  []byte
	Running bool // Whether the script is currently running.
}

// Collection contains all scripts.
type Collection struct {
	Scripts []Script
}

// Get creates a script collection.
func Get(dir string) Collection {
	c := Collection{make([]Script, 0, 16)}

	// Load all files that are executable.
	err := filepath.Walk(dir, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		// Check if info is an executable file.
		if !info.IsDir() && info.Mode().Perm()&0111 != 0 {
			var script Script
			script.ID = len(c.Scripts)
			script.Name = path.Base(filepath)
			script.path = filepath

			c.Scripts = append(c.Scripts, script)
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	return c
}

// Start the script
func (s *Script) Start() error {
	if s.Running {
		return errors.New("Script is already running.")
	}

	var cmd *exec.Cmd

	switch path.Ext(s.path) {
	case ".sh":
		cmd = exec.Command("/bin/sh", s.path)
	default:
		cmd = exec.Command("/bin/sh", s.path)
	}

	if cmd == nil {
		return errors.New("Could not create command")
	}

	stdoutIn, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	err := cmd.Start()
	if err != nil {
		return err
	}

	s.cmds = append(s.cmds, cmd)
	s.Running = true
	fmt.Println("is running")

	go func() {
		logfilename := strconv.Itoa(s.ID) + "_" + strconv.Itoa(len(s.cmds)-1) + ".log"
		fo, err := os.Create(logfilename)
		if err != nil {
			fmt.Println("Could not create " + logfilename + ": " + err.Error())
		} else {
			var out []byte
			buf := make([]byte, 8, 8)
			for {
				n, err := stdoutIn.Read(buf[:])
				if n > 0 {
					d := buf[:n]
					out = append(out, d...)
					_, err := fo.Write(d)
					if err != nil {
						fmt.Println(err)
					}
				}
				if err != nil {
					// Read returns io.EOF at the end of file, which is not an error for us
					if err != io.EOF {
						fmt.Println(err)
					}

					break
				}
			}
		}

		cmd.Wait()
		s.Running = false
		fmt.Println("finished")
	}()

	return nil
}

// Stop the script
func (s Script) Stop() error {
	cmd := s.cmds[len(s.cmds)-1]

	err := cmd.Process.Kill()
	if err != nil {
		return err
	}

	cmd.Wait()
	return nil
}

func (s Script) LastRunIndex() int {
	return len(s.cmds) - 1
}

func (s Script) Stdout(index int) (string, error) {
	if index >= len(s.cmds) {
		return "", errors.New("Invalid run index")
	}

	logfilename := strconv.Itoa(s.ID) + "_" + strconv.Itoa(index) + ".log"
	fo, err := os.Open(logfilename)

	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(fo)
	if err != nil {
		return "", nil
	}

	return string(data), nil
}
