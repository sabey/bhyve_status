# bhyve_status

This was originally written to check the status of bhyve under a pfSense host, but there's no reason why this shouldn't work on FreeBSD.

This script is meant to be run in any sort of shell script and will exit with a status code of 2 should any VM not be running or not responding. If an exit code of 2 is returned you should restart your VMs, or notify yourself that a VM is down.

- If a VM State is Running its process ID will be signaled with 0 to see if it is still running (and not hanging).
- If the State is anything but Running it's assumed to be down.

## Usage
```
-rc

This will load a list of VMs from an rc file, or any type of text configuration file can be specified in place of rc.
You may choose to use your own config file if you only want to watch for a certain set of VM IDs.
Rc/Config file must contain a line prefixed with this variable 'vm_list="' followed by IDs of VMs seperated by spaces, like so:

-rc vms.txt

$ cat vms.txt
vm_list="ubuntu windows"

-vm
This is how you get the VM status output from bhyve. Such as `/usr/local/sbin/vm list -v`
The expected output format is such:

$ ./vm_list.sh
NAME    DATASTORE  LOADER  CPU  MEMORY  VNC  AUTO      %CPU       RSZ          UPTIME  STATE
	ubuntu  default    grub    4    16G     -    Yes [1]    0.0    463.0M           07:27  Running (28159)
	windows default    grub    4    16G     -    Yes [1]      -         -               -  Stopped
	freebsd default    grub    4    16G     -    Yes [1]      -         -               -  Locked (router.pfsense.arpa)
	pfsense default    grub    4    16G     -    Yes [1]  Locked (router.pfsense.arpa)

```


## Usage (pfSense)
```
go build
./bhyve_status -rc rc.conf.local -vm ./vm_list.sh

nano /usr/local/etc/rc.d/vm

#!/bin/sh
#
# $FreeBSD$

# PROVIDE: vm
# REQUIRE: NETWORKING SERVERS dmesg
# BEFORE: dnsmasq ipfw pf
# KEYWORD: shutdown nojail

. /etc/rc.subr

: ${vm_enable="NO"}

name=vm
desc="Start and stop vm-bhyve guests on boot/shutdown"
rcvar=vm_enable

load_rc_config $name

command="/usr/local/sbin/${name}"
start_cmd="${name}_start"
stop_cmd="${command} stopall -f"
status_cmd="${name}_status"

vm_start()
{
    env rc_force="$rc_force" ${command} init
    env rc_force="$rc_force" ${command} startall >/dev/null &
}

vm_status()
{
  /opt/bhyve_status/bhyve_status
}

run_rc_command "$1"
```