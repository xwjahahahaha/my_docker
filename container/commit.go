package container

import (
	"os/exec"
	"xwj/mydocker/log"
)

// CommitContainer
// @Description: 打包一个容器
// @param imageName
func CommitContainer(imageName string) {
	mntUrl := "./mnt"
	imageTarUrl := "./" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTarUrl, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.LogErrorFrom("CommitContainer", "tar", err)
	}
}
