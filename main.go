package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

var (
	rc_conf_path string
	vm_list_cmd  string
)

func init() {
	flag.StringVar(&rc_conf_path, "rc_conf_path", "/etc/rc.conf.local", "rc conf path")
	flag.StringVar(&vm_list_cmd, "vm_list_cmd", "vm list -v", "vm list cmd")
}

const (
	rc_conf_vm_list_prefix = `vm_list="`
	rc_conf                = `# bhyve
	vm_enable="YES"
	vm_dir="zfs:pfsense/vm"
	vm_list="ubuntu"
	vm_delay="5"
`
	vm_list = `NAME    DATASTORE  LOADER  CPU  MEMORY  VNC  AUTO      %CPU       RSZ          UPTIME  STATE
	ubuntu  default    grub    4    16G     -    Yes [1]    0.0    463.0M           07:27  Running (28159)
	windows default    grub    4    16G     -    Yes [1]      -         -               -  Stopped
`
)

func main() {
	flag.Parse()

	watched := ScanRC()
	log.Printf("watched vms: %q\n", watched)

	allrunning := ScanVMS(watched)

	if !allrunning {
		log.Println("all watched vms are not running!")
		os.Exit(2)
	}
}

func ScanRC() []string {
	// scanner := bufio.NewScanner(bytes.NewBuffer([]byte(rc_conf)))

	f, err := os.Open(rc_conf_path)
	if err != nil {
		log.Fatalf("failed to open rc conf path: %q err: %q\n", rc_conf_path, err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, rc_conf_vm_list_prefix) {
			return strings.Fields(line[len(rc_conf_vm_list_prefix) : len(line)-1])
		}
	}

	return nil
}

func ScanVMS(
	watched []string,
) bool {
	allrunning := true

	// scanner := bufio.NewScanner(bytes.NewBuffer([]byte(vm_list)))

	result, err := exec.Command(vm_list_cmd).CombinedOutput()
	if err != nil {
		log.Fatalf("failed to execute vm list command: %q err: %q\n", vm_list_cmd, err)
		return false
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(result))
	for i := 0; scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		if i == 0 {
			if !strings.HasPrefix(line, "NAME") {
				log.Fatalf("vm header not found, got: %q\n", line)
				return false
			}
			log.Printf("%q", line)
			continue
		}
		if len(line) == 0 {
			continue
		}

		log.Printf("%q", line)

		row := strings.Fields(line)
		name := row[0]
		state := row[len(row)-1]
		running := !(state == "Stopped")
		pid := 0
		if running {
			if !strings.HasPrefix(state, "(") ||
				!strings.HasSuffix(state, ")") {
				log.Fatalf("running vm %q does not have a pid, got %q\n", name, line)
				return false
			}
			var err error
			pid, err = strconv.Atoi(state[1 : len(state)-1])
			if err != nil {
				log.Fatalf("running vm %q has invalid pid, got %q err: %q\n", name, line, err)
				return false
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				log.Printf("failed to find process vm %q pid %d err: %q", name, pid, err)
				if slices.Contains(watched, name) {
					log.Printf("watched vm %q is down!\n", name)
					allrunning = false
				} else {
					log.Printf("vm %q is down (not watched - ignored)\n", name)
				}
			} else {
				if err := process.Signal(syscall.Signal(0)); err != nil {
					log.Printf("failed to signal process vm %q pid %d err: %q", name, pid, err)
					if slices.Contains(watched, name) {
						log.Printf("watched vm %q is down!\n", name)
						allrunning = false
					} else {
						log.Printf("vm %q is down (not watched - ignored)\n", name)
					}
				}
				// running
			}
		} else {
			if slices.Contains(watched, name) {
				log.Printf("watched vm %q is not running!\n", name)
				allrunning = false
			} else {
				log.Printf("vm %q is not running (not watched - ignored)\n", name)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to scan... %q\n", err)
		return false
	}

	return allrunning
}
