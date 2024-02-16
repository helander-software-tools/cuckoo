package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
        "bufio"
	"io/ioutil"
)


func configSection(filePath string) []string {
    readFile, err := os.Open(filePath)
    if err != nil {
        fmt.Println(err)
	return nil
    }
    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)
    var fileLines []string
    for fileScanner.Scan() {
        line := strings.Trim(fileScanner.Text()," ")
	if len(line) > 0 {
          fileLines = append(fileLines, strings.Trim(fileScanner.Text()," "))
        }
    }
    readFile.Close()
    return fileLines
}


func main() {
        entrypointPtr := flag.String("entrypoint","", "entrypoint program")
	flag.Parse()
        if len(flag.Args()) < 1 {
	   fmt.Printf("\nMissing cuckoo/rootfs path\n")
	   os.Exit(1)
	}
	err:= os.Chdir(flag.Args()[0])
	if err != nil {
		fmt.Printf("\nCould not change directory to  %v\n",flag.Args()[0])
		os.Exit(1)
	}
	runCommand(flag.Args()[1:],*entrypointPtr)
}

func command(entrypoint []string, cmd []string, commandLine []string, entrypointflag string) []string {
        var args []string
	if len(entrypointflag) > 0 {
	  args = append(args,strings.Split(entrypointflag," ")...)
	} else {
	  if len(entrypoint) > 0 {
	    args = append(args,entrypoint...)
	  } else {
	    args = append(args,"/bin/sh", "-c")
          }
	}

        var command  []string
	if len(cmd) > 0 {
	  command = append(command,cmd...)
	}
	if len(commandLine) > 0 {
	  command = append(command,commandLine...)
	}

	args = append(args,command...)
	return args
}

func runCommand(args []string, entrypoint string) {

        var Dir []string
        var Env []string
	var Entrypoint[]string
	var Cmd[]string
	var progCmd[]string
	
        Cmd = configSection(".cuckoo/cmd")
	Entrypoint = configSection(".cuckoo/entrypoint")
        Env = configSection(".cuckoo/env")
        Dir = configSection(".cuckoo/dir")

        if len(args) < 1 {
	  progCmd = command(Entrypoint,Cmd,[]string{},entrypoint)
	} else {
	  progCmd = command(Entrypoint,Cmd,args,entrypoint)
        }
	//fmt.Printf("COMMAND  %v",progCmd)
	cmd := exec.Command(progCmd[0], progCmd[1:]...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Env = Env

	bytesRead, err := ioutil.ReadFile("/etc/resolv.conf")
        if err != nil {
            panic(err)
        }
        err = ioutil.WriteFile("etc/resolv.conf", bytesRead, 0644)
        if err != nil {
          panic(err)
        }

        /* In case we need individual mount points for each container, create a tempdir somewhere and mount on subdirs
            replace the defer method with a function that does umounts and after that os.RemoveAll 
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up
        */

	must(syscall.Mount("/proc", "proc", "", syscall.MS_BIND, ""))
        must(syscall.Mount("/dev", "dev", "", syscall.MS_BIND, ""))
        must(syscall.Mount("/sys", "sys", "", syscall.MS_BIND, ""))

	must(syscall.Chroot("."))
	must(os.Chdir("/"))
	
	if len(Dir) > 0  {
	   os.Chdir(Dir[0])
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
	   sig := <-sigs
	   fmt.Println()
	   fmt.Println(sig)
	}()

	err = cmd.Run()
        if err != nil {
		fmt.Printf("\nError result from program : %v\n",err)
        }

        must(syscall.Unmount("/dev", 0))
        must(syscall.Unmount("/sys", 0))
        must(syscall.Unmount("/proc", 0))

	fmt.Printf("\n")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
