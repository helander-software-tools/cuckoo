package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"io/ioutil"
)



func jsonStringConfig(filePath string) string {
    b, err := os.ReadFile(filePath)
    if err != nil {
        fmt.Printf("Unable to read file due to %s\n", err)
	return ""
    }

    var result string

    err = json.Unmarshal(b, &result)
    if err != nil {
        fmt.Printf("Unable to marshal JSON due to %s", err)
	return ""
    }
    return result

}

func jsonArrayConfig(filePath string) []string {
    b, err := os.ReadFile(filePath)
    if err != nil {
        fmt.Printf("Unable to read file due to %s\n", err)
	return nil
    }

    var result []string

    err = json.Unmarshal(b, &result)
    if err != nil {
        fmt.Printf("Unable to marshal JSON due to %s", err)
	return nil
    }
    return result

}

func main() {
	flag.Usage = func() {
	  fmt.Fprintf(flag.CommandLine.Output(), "Usage:   %s [flags] rootfs-dir [arguments]\n", os.Args[0])
	  flag.PrintDefaults()
	}
        entrypointPtr := flag.String("entrypoint","", "entrypoint program")
	flag.Parse()
        if len(flag.Args()) < 1 {
	   fmt.Printf("\nMissing cuckoo/rootfs path\n")
	   flag.Usage()
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
	  args = append(args,entrypoint...)
	}

	if len(commandLine) > 0 {
	  args = append(args,commandLine...)
	} else {
	  args = append(args,cmd...)
	}

	return args
}

func runCommand(args []string, entrypoint string) {

        var Dir string
        var Env []string
	var Entrypoint[]string
	var Cmd[]string
	var progCmd[]string
	
        Cmd = jsonArrayConfig(".cuckoo/cmd")
	Entrypoint = jsonArrayConfig(".cuckoo/entrypoint")
        Env = jsonArrayConfig(".cuckoo/env")
        Dir = jsonStringConfig(".cuckoo/dir")

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
	   os.Chdir(Dir)
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
