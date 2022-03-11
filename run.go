package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
        "bufio"
	"io/ioutil"
)


func configSection(filePath string) []string {
 
    readFile, err := os.Open(filePath)
    //fmt.Printf("\nconfigSection %s",filePath)
  
    if err != nil {
        fmt.Println(err)
	return nil
    }
    fileScanner := bufio.NewScanner(readFile)
    fileScanner.Split(bufio.ScanLines)
    var fileLines []string
  
    for fileScanner.Scan() {
        fileLines = append(fileLines, fileScanner.Text())
    }
  
    readFile.Close()
  
    return fileLines

}


func main() {
	if len(os.Args) < 2 {
	   fmt.Printf("\nMissing command.")
	   os.Exit(1)
	}
	switch os.Args[1] {
	case "run":
		runCommand(os.Args[2:])
	default:
		fmt.Printf("\nmain() with args %v \n",os.Args)
		panic("what should I do")
	}
}

	//cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)


func command(entrypoint []string, cmd []string, commandLine []string) []string {
        var args []string
	
	fmt.Printf("\n   entrypoint:  %v",entrypoint)
	fmt.Printf("\n   cmd:         %v",cmd)
	fmt.Printf("\n   commandLine: %v",commandLine)

	args = commandLine
	return args
}

func runCommand(args []string) {
        if len(args) < 1 {
	   fmt.Printf("\nMissing rootfs path")
	   os.Exit(1)
	}
	must(os.Chdir(args[0]))	
	

        var Dir []string
        var Env []string
	var Entrypoint[]string
	var Cmd[]string
	
        Cmd = configSection(".cuckoo/cmd")     
	fmt.Printf("\nCmd[] %v",Cmd)
	Entrypoint = configSection(".cuckoo/entrypoint")
	fmt.Printf("\nEntrypoint[] %v",Entrypoint)
        Env = configSection(".cuckoo/env")     
	fmt.Printf("\nEnv[] %v",Env)
        Dir = configSection(".cuckoo/dir")   
	fmt.Printf("\nDir[] %v",Dir)
	
	var progCmd []string := command(Entrypoint,Cmd,args[1:])
	fmt.Printf("\nProgram exec : %v",progCmd)
	
	cmd := exec.Command(progCmd)
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


	must(syscall.Mount("/proc", "proc", "", syscall.MS_BIND, ""))
        must(syscall.Mount("/dev", "dev", "", syscall.MS_BIND, ""))    
        must(syscall.Mount("/sys", "sys", "", syscall.MS_BIND, ""))
	
	must(syscall.Chroot("."))
	must(os.Chdir("/"))	
	
	if Dir[0] != "" {
	   fmt.Printf("\nChange directory to: %s",Dir[0])
	   os.Chdir(Dir[0])
	}
	fmt.Printf("\n\nRun program:\n\n")
	err = cmd.Run()
        if err != nil {
		fmt.Printf("\nError result from program : %v\n",err)
        }
	
        must(syscall.Unmount("/dev", 0))
        must(syscall.Unmount("/sys", 0))
        must(syscall.Unmount("/proc", 0))
	
	fmt.Printf("\nExit program.\n")
}  

func must(err error) {
	if err != nil {
		//fmt.Printf("\n\nError must %v",err)
		panic(err)
	}
}

