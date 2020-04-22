package nsenter

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void  enter_namespace(void)  {
	char *tinydocker_pid = "";
	tinydocker_pid = getenv("tinydocker_pid");
	//int log_fd = open("/var/run/tinydocker/nsenter.log", O_WRONLY | O_CREAT | O_TRUNC, 0644);
	if (!tinydocker_pid) {
	    fprintf(stdout, "fail to get pid of container \n");
		return;
	}

	char *tinydocker_command = "";
	tinydocker_command = getenv("tinydocker_command");
	if (!tinydocker_command) {
		fprintf(stdout, "fail to get command of container \n");
		return;
	}

	int i = 0;
	char nspath[1024];
	char *namespace[] = {"ipc", "uts", "pid", "net", "mnt"};
	for (i = 0; i < 5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", tinydocker_pid, namespace[i]);
		int fd = open(nspath, O_RDONLY);
		if (setns(fd, 0) == -1) {
			fprintf(stderr, "setns on %s namespace error : %s\n", namespace[i], strerror(errno));
		}
		close(fd);
	}
	int res = system(tinydocker_command);
	exit(0);
	return;
}
*/
import "C"
