# bhyve_status
```
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