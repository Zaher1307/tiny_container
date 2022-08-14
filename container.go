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
        err := run()
		if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "child":
        err := child()
		if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid arguments")
		os.Exit(1)
	}
}

func run() error {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Unshareflags: syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      syscall.Getuid(),
			Size:        1,
		}},
		GidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      syscall.Getgid(),
			Size:        1,
		}},
		Credential: &syscall.Credential{
			Uid: uint32(syscall.Getuid()),
			Gid: uint32(syscall.Getuid()),
		},
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func child() error {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cg()
	if err != nil {
		return err
	}

	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
	err = syscall.Chroot("./root-fs")
	if err != nil {
		return err
	}
	err = syscall.Chdir("/")
	if err != nil {
		return err
	}
	err = syscall.Mount("proc", "proc", "proc", 0, "")
	if err != nil {
		return err
	}
	err = syscall.Mount("dev", "dev", "tmpfs", 0, "")
	if err != nil {
		return err
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	err = syscall.Unmount("proc", 0)
	if err != nil {
		return err
	}
	err = syscall.Unmount("dev", 0)
	if err != nil {
		return err
	}

	return nil
}

func cg() error {
	cgroups := "/sys/fs/cgroup/"
	pids := path.Join(cgroups, "pids")
	os.Mkdir(pids, 0755)
	err := os.WriteFile(path.Join(pids, "pids.max"), []byte("20"), 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(pids, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
	if err != nil {
		return err
	}
	return nil
}
