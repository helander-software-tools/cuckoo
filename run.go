package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
        "bufio"
)

func main()  {
        var Args []string
        var Dir []string
        var Env []string
        Args = configSection(".cuckoo/args")     
	fmt.Printf("\nArgs[] %v",Args)
        Env = configSection(".cuckoo/env")     
	fmt.Printf("\nEnv[] %v",Env)
        Dir = configSection(".cuckoo/dir")   
	fmt.Printf("\nDir[] %v",Dir)
        cmd := exec.Command(Args[0], Args[1:]...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Env = Env
        cmd.SysProcAttr = &syscall.SysProcAttr{
                Cloneflags: syscall.CLONE_NEWUTS |
                        syscall.CLONE_NEWPID |
                        syscall.CLONE_NEWNS |
                        syscall.CLONE_NEWIPC,
        }
	fmt.Printf("\nBefore chroot")
	syscall.Chroot(".")
	fmt.Printf("\nBefore chdir")
	os.Chdir(Dir[0])
	fmt.Printf("\nBefore run")
	err := cmd.Run()
        if err != nil {
                log.Fatal(err)
        }
}

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



