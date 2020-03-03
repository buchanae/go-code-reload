package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"
	"time"
)

func main() {
	log.SetFlags(0)
	target := "plugin.go"

	// Rebuild and reload the plugin every second.
	for {

		// Create temp dir for building the plugin.
		dir, err := ioutil.TempDir("", "code-reload-")
		if err != nil {
			panic(err)
		}

		// Copy the plugin file to the build directory.
		codePath := filepath.Join(dir, filepath.Base(target))
		err = copyFile(codePath, target)
		if err != nil {
			panic(err)
		}
		// Create a unique plugin identity by creating a unique module name.
		modContent := fmt.Sprintf("module code_reload_%s\n", time.Now().Unix())

		// Write the go.mod file to the build directory.
		modPath := filepath.Join(dir, "go.mod")
		err = ioutil.WriteFile(modPath, []byte(modContent), 0644)
		if err != nil {
			panic(err)
		}

		err = buildPlugin(dir)
		if err != nil {
			panic(err)
		}

		// Load the plugin
		pluginPath := filepath.Join(dir, "code_reload")
		p, err := plugin.Open(pluginPath)
		if err != nil {
			panic(err)
		}
		sym, err := p.Lookup("Name")
		if err != nil {
			panic(err)
		}
		Name := sym.(func() string)

		log.Printf("Name: %s", Name())

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		log.Printf("memory usage: %d", m1.Alloc)

		os.RemoveAll(dir)
		//log.Print(dir)
		//break

		time.Sleep(time.Second)
	}

}

func buildPlugin(path string) error {
	cmd := exec.Command(
		"go", "build", "-buildmode=plugin", "-o=code_reload", ".",
	)
	cmd.Env = []string{}
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = path

	return cmd.Run()
}

func copyFile(dstPath, srcPath string) error {
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	defer src.Close()
	return nil
}
