//+build linux
package namespace

/*
#define _GNU_SOURCE
#include <fcntl.h>
#include <sched.h>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>

// _attribute_((constructor))指的是一旦这个包被调用那么这个函数就会自动被执行。类似于构造函数，会在程序一启动的时候运行
__attribute__((constructor)) void enter_namespace(void) {
	char *mydocker_pid;
	// 从环境变量中获取需要进入的PID
	mydocker_pid = getenv("mydocker_pid");
	if (mydocker_pid) {
		fprintf(stdout, "C : mydocker_pid = %s\n", mydocker_pid);
	}else {
		//这里如果没有指定pid，那么就不需要继续向下执行了
		return;
	}
	char *mydocker_cmd;
	// 从环境变量中获取需要执行的命令
	mydocker_cmd = getenv("mydocker_cmd");
	if (mydocker_cmd) {
		fprintf(stdout, "C : mydocker_cmd = %s\n", mydocker_cmd);
	}else {
		// 同理
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt"};
	// 循环每一个Namespace，让进程进入
	for (i=0; i<5; i++) {
		// 拼接对应的路径/proc/pid/ns/ipc
		sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);
		// 调用setns系统调用实现进入对应的Namespace, -1是成功的返回值
		if (setns(fd, 0) == -1) {
			return;
		}
		close(fd);
	}
	// 进入所有Namespace后执行指定的命令
	int res = system(mydocker_cmd);
	exit(0);
	return;
}
*/
import "C"

func EnterNamespace()  {}
