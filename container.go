package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

func main() {
    switch os.Args[1] {
    case "run":
        run()
    case "child":
        child()
    default:
        panic("bad command")
    }
}

func run() {
    fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID: syscall.Getuid(),
			Size: 1,
		}},
		GidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID: syscall.Getgid(),
			Size: 1,
		}},
		Credential: &syscall.Credential{
			Uid: uint32(syscall.Getuid()),
			Gid: uint32(syscall.Getuid()),
		},
	}

    must(cmd.Run())
}

func child() {
    fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

    cmd := exec.Command(os.Args[2], os.Args[3:]...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    cg()

    must(syscall.Sethostname([]byte("container")))
    must(syscall.Chroot("./alpine-fs"))
    must(syscall.Chdir("/"))
    must(syscall.Mount("proc", "proc", "proc", 0, ""))
    must(syscall.Mount("dev", "dev", "tmpfs", 0, ""))

    must(cmd.Run())

    must(syscall.Unmount("proc", 0))
    must(syscall.Unmount("dev", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := path.Join(cgroups, "pids")
	os.Mkdir(pids, 0755)
	must(os.WriteFile(path.Join(pids, "pids.max"), []byte("20"), 0700))
	must(os.WriteFile(path.Join(pids, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
    if err != nil {
        panic(err)
    }
}
