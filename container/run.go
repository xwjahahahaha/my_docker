package container

import "xwj/mydocker/log"

// Run
// @Description: 运行docker
// @param tty
// @param cmd
func Run(tty bool, cmd string){
	log.Log.Infof("call Run tty = %v, cmd = %s", tty, cmd)
}
