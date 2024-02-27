package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"io/ioutil"
)


func main() {
	flag.Usage = func() {
	  fmt.Fprintf(flag.CommandLine.Output(), "Usage:   %s create|run|exec|rm [flags] container-path [arguments]\n", os.Args[0])
	  flag.PrintDefaults()
	}


        createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createCmd.Usage = func() {
	  fmt.Printf("Usage: %s  create container-path image-path\n",os.Args[0])
	  createCmd.PrintDefaults()
	}

        rmCmd := flag.NewFlagSet("rm", flag.ExitOnError)
	rmCmd.Usage = func() {
	  fmt.Printf("Usage: %s  rm container-path\n",os.Args[0])
	  rmCmd.PrintDefaults()
	}

        execCmd := flag.NewFlagSet("exec", flag.ExitOnError)
	execCmd.Usage = func() {
	  fmt.Printf("Usage: %s  exec container-path [command [arguments]]\n",os.Args[0])
	  execCmd.PrintDefaults()
	}

        runCmd := flag.NewFlagSet("run", flag.ExitOnError)
        runEntrypoint := runCmd.String("entrypoint","", "entrypoint program")
        runOutputpath := runCmd.String("output","", "file for stdout/stderr")
        runVerbose := runCmd.Bool("verbose",false, "verbose output")
	runCmd.Usage = func() {
	  fmt.Printf("Usage: %s  run [flags] container-path [command [arguments]]\n",os.Args[0])
	  runCmd.PrintDefaults()
	}


        if len(os.Args) < 2 {
          fmt.Println("expected subcommand")
          flag.Usage()
          os.Exit(1)
        }

        switch os.Args[1] {

          case "create":
             createCmd.Parse(os.Args[2:])
             if len(createCmd.Args()) < 1 {
     	       fmt.Printf("\nMissing container path\n")
	       createCmd.Usage()
	       os.Exit(1)
	     }
             createCommand(createCmd.Args()[0],createCmd.Args()[1:])
          case "rm":
             rmCmd.Parse(os.Args[2:])
             if len(rmCmd.Args()) < 1 {
     	       fmt.Printf("\nMissing container path\n")
	       rmCmd.Usage()
	       os.Exit(1)
	     }
             rmCommand(rmCmd.Args()[0],rmCmd.Args()[1:])
          case "exec":
             execCmd.Parse(os.Args[2:])
             if len(execCmd.Args()) < 1 {
     	       fmt.Printf("\nMissing container path\n")
	       execCmd.Usage()
	       os.Exit(1)
	     }
             execCommand(execCmd.Args()[0],execCmd.Args()[1:])
          case "run":
             runCmd.Parse(os.Args[2:])
             if len(runCmd.Args()) < 1 {
     	       fmt.Printf("\nMissing container path\n")
	       runCmd.Usage()
	       os.Exit(1)
	     }
             runCommand(runCmd.Args()[0],runCmd.Args()[1:],*runEntrypoint,*runVerbose,*runOutputpath)
          default:
             fmt.Println("unknown subcommand")
             flag.Usage()
             os.Exit(1)
        }
}

func IsDirEmpty(name string) (bool) {
         f, err := os.Open(name)
         if err != nil {
                 return false
         }
         defer f.Close()
         _, err = f.Readdir(1)
         if err == io.EOF {
                 return true
         }
         return false
}





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

func execCommand(containerpath string,args []string) {

        var Dir string
        var Env []string

	rootfspath := filepath.Join(containerpath,"rootfs")

        Env = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/env"))
        Dir = jsonStringConfig(filepath.Join(rootfspath,".cuckoo/dir"))

	var cmd *exec.Cmd
	if len(args) < 2 {
	  cmd = exec.Command(args[0])
        } else {
	  cmd = exec.Command(args[0], args[1:]...)
	}
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Env = Env
	err:= os.Chdir(rootfspath)
	if err != nil {
		fmt.Printf("\nCould not change directory to  %v\n",rootfspath)
		os.Exit(1)
	}

	must(syscall.Chroot("."))

        if IsDirEmpty("/proc") {
		fmt.Printf("\nNot an active container\n")
		os.Exit(1)
	}
	if len(Dir) > 0  {
	   must(os.Chdir(Dir))
	} else {
	   must(os.Chdir("/"))
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

	err = cmd.Run()
        if err != nil {
		fmt.Printf("\nError result from program : %v\n",err)
        }
}



func runCommand(containerpath string,args []string, entrypoint string, verbose bool,outputpath string) {

        var Dir string
        var Env []string
	var Entrypoint[]string
	var Cmd[]string
	var progCmd[]string
	
	rootfspath := filepath.Join(containerpath,"rootfs")

        Cmd = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/cmd"))
	Entrypoint = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/entrypoint"))
        Env = jsonArrayConfig(filepath.Join(rootfspath,".cuckoo/env"))
        Dir = jsonStringConfig(filepath.Join(rootfspath,".cuckoo/dir"))

	progCmd = command(Entrypoint,Cmd,args,entrypoint)

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

        if IsDirEmpty("proc") {
          must(syscall.Mount("/proc", "proc", "", syscall.MS_BIND, ""))
          must(syscall.Mount("/dev", "dev", "", syscall.MS_BIND, ""))
          must(syscall.Mount("/dev/pts", "dev/pts", "", syscall.MS_BIND, ""))
          must(syscall.Mount("/sys", "sys", "", syscall.MS_BIND, ""))
	}

	must(syscall.Chroot("."))
	must(os.Chdir("/"))

	if len(Dir) > 0  {
	   os.Chdir(Dir)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

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


}

func createCommand(containerpath string,args []string) {
	
	rootfspath := filepath.Join(containerpath,"rootfs")

/*
Here we shall mount the squasfs image and mount the overlay. We also shall create all dirs and links under the containerpath.

*/

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
          must(syscall.Mount("/dev/pts", "dev/pts", "", syscall.MS_BIND, ""))
          must(syscall.Mount("/sys", "sys", "", syscall.MS_BIND, ""))

}

func rmCommand(containerpath string,args []string) {
	
	rootfspath := filepath.Join(containerpath,"rootfs")


	err:= os.Chdir(rootfspath)
	if err != nil {
		fmt.Printf("\nCould not change directory to  %v\n",rootfspath)
		os.Exit(1)
	}


          syscall.Unmount("sys", syscall.MNT_FORCE)
          syscall.Unmount("dev/pts", syscall.MNT_FORCE)
          syscall.Unmount("dev", syscall.MNT_FORCE)
          syscall.Unmount("proc", syscall.MNT_FORCE)

	err = os.Chdir("..")
          syscall.Unmount("rootfs", syscall.MNT_FORCE)
          syscall.Unmount("image", syscall.MNT_FORCE)


/*
Here we shall delete all dirs and links under and at  the containerpath.
*/

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

