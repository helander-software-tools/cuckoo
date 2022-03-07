package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
        "bufio"
)


func configSection(filePath string) []string {
 
    readFile, err := os.Open(filePath)
    fmt.Printf("\nconfigSection %s",filePath)
  
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
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("what should I do")
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func child() {

	
	var Args []string
        var Dir []string
        var Env []string
        Args = configSection("rootfs/.cuckoo/args")     
	fmt.Printf("\nArgs[] %v",Args)
        Env = configSection("rootfs/.cuckoo/env")     
	fmt.Printf("\nEnv[] %v",Env)
        Dir = configSection("rootfs/.cuckoo/dir")   
	fmt.Printf("\nDir[] %v",Dir)
	if len(os.Args) > 2 {
	   fmt.Printf("\nos.Args %v",os.Args[2:])
	   Args = os.Args[2:]
	   fmt.Printf("\nArgs[] %v",Args)
	}
	
	must(syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""))
	must(os.MkdirAll("rootfs/oldrootfs", 0700))
	must(syscall.PivotRoot("rootfs", "rootfs/oldrootfs"))
	must(os.Chdir("/"))
	
	
        cmd := exec.Command(Args[0], Args[1:]...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Env = Env

	if Dir[0] != "" {
	   fmt.Printf("\nBefore chdir")
	   os.Chdir(Dir[0])
	}
	fmt.Printf("\nBefore run")
	err := cmd.Run()
        if err != nil {
                log.Fatal(err)
        }
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

