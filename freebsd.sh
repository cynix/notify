#!/bin/sh
#
# PROVIDE: notify
# REQUIRE: NETWORKING
# KEYWORD:

. /etc/rc.subr

name="notify"
rcvar="notify_enable"

load_rc_config $name

: ${notify_enable:="NO"}
: ${notify_username:="mailnull"}
: ${notify_logfile:="/var/log/notify.log"}

pidfile="/var/run/notify.pid"
command="/usr/sbin/daemon"
command_args="-o ${notify_logfile} -P ${pidfile} -r -u ${notify_username} /usr/local/bin/notify --daemon"
required_files="/usr/local/etc/notify.conf"
start_precmd="mkdir -p /var/run/notify && chown ${notify_username} /var/run/notify && chmod 755 /var/run/notify"
stop_postcmd="rm -f /var/run/notify/sock"

run_rc_command "$1"
