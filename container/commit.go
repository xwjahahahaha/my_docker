package container

import (
	"os/exec"
	"path/filepath"
	"xwj/mydocker/log"
)

// CommitContainer
// @Description: 打包一个容器
// @param imageName
func CommitContainer(containerID, imageName string) {
	mntUrl := filepath.Join(ROOTURL, "mnt", containerID)
	imageTarUrl := "./" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTarUrl, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.LogErrorFrom("CommitContainer", "tar", err)
	}
}
