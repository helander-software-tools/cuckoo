package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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
        outputpathPtr := flag.String("output","", "file for stdout/stderr")
        verbosePtr := flag.Bool("verbose",false, "verbose output")
	flag.Parse()
        if len(flag.Args()) < 1 {
	   fmt.Printf("\nMissing cuckoo/rootfs path\n")
	   flag.Usage()
	   os.Exit(1)
	}
	runCommand(flag.Args()[0],flag.Args()[1:],*entrypointPtr,*verbosePtr,*outputpathPtr)
}

func command(entrypoint []string, cmd []string, commandLine []string, entrypointflag string) []string {
        var args []string

	if len(entrypointflag) > 0 {
	  args = append(args,entrypointflag)
	} else {
	  args = append(args,entrypoint...)
	}

	if len(commandLine) > 0 {
	  args = append(args,commandLine...)
	} else {
          // This seems to be the logic for docker run --entrypoint
	  if len(entrypointflag) < 1 {
	    args = append(args,cmd...)
	  }
	}

	return args
}

func runCommand(rootfspath string,args []string, entrypoint string, verbose bool,outputpath string) {

        var Dir string
        var Env []string
	var Entrypoint[]string
	var Cmd[]string
	var progCmd[]string
	
        Cmd = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/cmd"))
	Entrypoint = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/entrypoint"))
        Env = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/env"))
        Dir = jsonStringConfig(filepath.Join(rootfspath,".cuckoo/dir"))

	//fmt.Printf("ARGS %v\n",args)
        //if len(args) < 1 {
	//  progCmd = command(Entrypoint,Cmd,[]string{},entrypoint)
	//} else {
	  progCmd = command(Entrypoint,Cmd,args,entrypoint)
        //}
        if verbose {fmt.Printf("Run command   %v\n",progCmd)}
	var cmd *exec.Cmd
	if len(progCmd) < 2 {
	  cmd = exec.Command(progCmd[0])
        } else {
	  cmd = exec.Command(progCmd[0], progCmd[1:]...)
	}
        cmd.Stdin = os.Stdin
	if len(outputpath) > 0 {
	  outfile, err := os.Create(outputpath)
 	  if err != nil {
	    panic(err)
	  }
	  writer := bufio.NewWriter(outfile)
	  defer writer.Flush()
	  defer outfile.Close()
	  cmd.Stdout = outfile
          cmd.Stderr = outfile
	} else {
          cmd.Stdout = os.Stdout
          cmd.Stderr = os.Stderr
	}
        cmd.Env = Env
	err:= os.Chdir(rootfspath)
	if err != nil {
		fmt.Printf("\nCould not change directory to  %v\n",rootfspath)
		os.Exit(1)
	}

	bytesRead, err := ioutil.ReadFile("/etc/resolv.conf")
        if err != nil {
            panic(err)
        }
        err = ioutil.WriteFile("etc/resolv.conf", bytesRead, 0644)
        if err != nil {
          panic(err)
        }

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


	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

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
