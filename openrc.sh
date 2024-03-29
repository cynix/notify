#!/sbin/openrc-run

name="Notify"
command="/usr/local/bin/notify"
command_args="--daemon"
command_background="yes"
command_user="mail"
pidfile="/run/$RC_SVCNAME.pid"
output_log="/var/log/notify.log"
error_log="/var/log/notify.err"
required_files="/usr/local/etc/notify.conf"

depend() {
	need network-online
	use dns logger
}

start_pre() {
	checkpath --directory --mode 0755 --owner $command_user /var/run/notify
}

stop_post() {
	rm -f /var/run/notify/sock
}
