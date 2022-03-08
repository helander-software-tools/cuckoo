package main

import (
	"fmt"
	"log"
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
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		fmt.Printf("\nmain() with args %v \n",os.Args)
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

	must(cmd.Run())
}

func child() {

	
	
	var Args []string
        var Dir []string
        var Env []string
        Args = configSection(".cuckoo/args")     
	fmt.Printf("\nArgs[] %v",Args)
        Env = configSection(".cuckoo/env")     
	fmt.Printf("\nEnv[] %v",Env)
        Dir = configSection(".cuckoo/dir")   
	fmt.Printf("\nDir[] %v",Dir)
	if len(os.Args) > 2 {
	   //fmt.Printf("\nos.Args %v",os.Args[2:])
	   Args = os.Args[2:]
	   fmt.Printf("\nArgs[] %v",Args)
	}

	
        cmd := exec.Command(Args[0], Args[1:]...)
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

	must(syscall.Chroot("."))
	must(os.Chdir("/"))	

	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
        must(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""))
        must(syscall.Mount("tmpfs", "/dev", "tmpfs", 0, ""))    
	must(os.MkdirAll("/dev/pts",0755))
        must(syscall.Mount("devpts", "/dev/pts", "devpts", 0, ""))
        must(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""))
	
	
	if Dir[0] != "" {
	   fmt.Printf("\nChange directory to: %s",Dir[0])
	   os.Chdir(Dir[0])
	}
	fmt.Printf("\nRun container:\n")
	err = cmd.Run()
        if err != nil {
                log.Fatal(err)
        }
	
	must(syscall.Unmount("/dev/pts", 0))
        must(syscall.Unmount("/dev", 0))
        must(syscall.Unmount("/sys", 0))
        must(syscall.Unmount("/proc", 0))
        must(syscall.Unmount("/tmp", 0))
}  

func must(err error) {
	if err != nil {
		//fmt.Printf("\n\nError must %v",err)
		panic(err)
	}
}

