package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main()  {
        var Args []string
        var Dir []string
        var Env []string
        Args = configSection(".cuckoo/args")        
        Env = configSection(".cuckoo/env")        
        Dir = configSection(".cuckoo/dir")        
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
	syscall.Chroot(".")
	os.Chdir(Dir[0])
	err := cmd.Run()
        if err != nil {
                log.Fatal(err)
        }
}

func configSection(filePath fp) []string {
 
    readFile, err := os.Open(filePath)
  
    if err != nil {
        fmt.Println(err)
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



