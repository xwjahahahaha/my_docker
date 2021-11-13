package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xwj/mydocker/log"
)

// NewWorkSpace
// @Description: 创建新的文件工作空间
// @param rootURL
// @param mntURL
// @param volume 是否使用数据卷
func NewWorkSpace(rootURL, imageName, mntURL, volume string) {
	CreateReadOnlyLayer(rootURL, imageName)      // 创建init只读层
	CreateWriteLayer(rootURL)                    // 创建读写层
	CreateMountPoint(rootURL, imageName, mntURL) // 创建mnt文件夹并挂载
	if volume != "" {
		// 数据卷操作
		volumeUrls, err := volumeUrlExtract(volume)
		if err != nil {
			log.Log.Warn(err)
			return
		}
		// 挂载Volume
		MountVolume(mntURL, volumeUrls)
		log.Log.Infof("success establish volume : %s", strings.Join(volumeUrls, ""))
	}
}

// volumeUrlExtract
// @Description: 解析volume字符串
// @param volume
// @return []string
func volumeUrlExtract(volume string) ([]string, error)  {
	volumeAry := strings.Split(volume, ":")
	if len(volumeAry) != 2 || volumeAry[0] == "" || volumeAry[1] == "" {
		return nil, fmt.Errorf(" Invalid volume string!")
	}
	return volumeAry, nil
}


// MountVolume
// @Description: 挂载数据卷
// @param mntUrl
// @param volumeUrl
func MountVolume(mntUrl string, volumeUrl []string)  {
	// 1. 创建宿主机文件目录
	parentUrl, containerUrl := volumeUrl[0], filepath.Join(mntUrl, volumeUrl[1])
	if has, err := dirOrFileExist(parentUrl); err == nil && !has {
		// 当宿主机没有此文件时，创建文件夹
		if err := os.Mkdir(parentUrl, 0777); err != nil {
			log.LogErrorFrom("MountVolume", "Mkdir", err)
			return
		}
	}
	// 2. 在容器目录中创建挂载点目录
	if has, err := dirOrFileExist(containerUrl); err == nil && has {
		// 如果有此文件夹，则先删除
		if err := os.RemoveAll(containerUrl); err != nil {
			log.LogErrorFrom("MountVolume", "RemoveAll", err)
			return
		}
	}
	// 容器中创建文件夹
	if err := os.Mkdir(containerUrl, 0777); err != nil {
		log.LogErrorFrom("MountVolume", "Mkdir", err)
		return
	}
	// 3. 将宿主机的文件目录挂载到容器挂载点
	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "myDockerVolume", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Log.Errorf("Mount volume failed. %v", err)
	}
}


// CreateReadOnlyLayer
// @Description: 通过镜像的压缩包解压并创建镜像文件夹作为只读层
// @param rootURL
// @param imageName
func CreateReadOnlyLayer(rootURL, imageName string) {
	imageName = strings.Trim(imageName, "/")
	imageDir := rootURL + imageName + "/"
	imageTarPath := rootURL + imageName + ".tar"
	if has, err := dirOrFileExist(imageTarPath); err == nil && !has {
		log.Log.Errorf(" Target image tar file not exist!")
		return
	}
	if has, err := dirOrFileExist(imageDir); err == nil && !has {
		// 创建文件夹
		if err := os.Mkdir(imageDir, 0777); err != nil {
			log.LogErrorFrom("createReadOnlyLayer", "Mkdir", err)
		}
	}
	if _, err := exec.Command("tar", "-xvf", imageTarPath, "-C", imageDir).CombinedOutput(); err != nil {
		log.LogErrorFrom("createReadOnlyLayer", "tar", err)
	}
}

// CreateWriteLayer
// @Description: 创建读写层
// @param rootURL
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if has, err := dirOrFileExist(writeURL); err == nil && has {
		log.Log.Info("Write layer dir already exist. Delete and create new one.")
		// 如果存在则先删除掉之前的
		DeleteWriteLayer(rootURL)
	}
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.LogErrorFrom("createWriteLayer", "Mkdir", err)
	}
}

// CreateMountPoint
// @Description: 挂载到容器目录mnt
// @param rootURL
// @param imageName
// @param mntURL
func CreateMountPoint(rootURL, imageName, mntURL string) {
	if has, err := dirOrFileExist(mntURL); err == nil && has {
		log.Log.Info("mnt dir already exist. Delete and create new one.")
		DeleteMountPoint(mntURL)
	}
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.LogErrorFrom("CreateMountPoint", "Mkdir", err)
	}
	// 将读写层目录与镜像只读层目录mount到mnt目录下
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + imageName
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "myDockerMnt", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("createMountPoint", "mount", err)
	}
}

// DeleteWorkSpace
// @Description: 当容器删除时一起删除工作空间
// @param rootURL
// @param mntURL
func DeleteWorkSpace(rootURL, mntURL, volume string) {
	if volume != "" {
		// 当volume不为空的时候，
		volumeUrls, err := volumeUrlExtract(volume);
		if err != nil {
			// 解析错误
			log.Log.Warn(err)
			DeleteMountPoint(mntURL)
		}else {
			DeleteMountPointWithVolume(mntURL, volumeUrls)
		}
	}else {
		DeleteMountPoint(mntURL)
	}
	DeleteWriteLayer(rootURL)
}

// DeleteMountPoint
// @Description: 取消挂载点并删除mnt目录
// @param mntURL
func DeleteMountPoint(mntURL string) {
	// 取消mnt目录的挂载
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("deleteMountPoint", "umount", err)
	}
	// 删除mnt目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.LogErrorFrom("deleteMountPoint", "remove", err)
	}
}

// DeleteWriteLayer
// @Description: 删除读写层目录
// @param rootURL
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.LogErrorFrom("deleteWriteLayer", "remove", err)
	}
}

// DeleteMountPointWithVolume
// @Description: 卸载volume的挂载点并删除容器挂载层
// @param mntURL
// @param volumeUrls
func DeleteMountPointWithVolume(mntURL string, volumeUrls []string)  {
	// 其实相比而言就是多了一步卸载volume的挂载点
	containerUrl := filepath.Join(mntURL, volumeUrls[1])
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("deleteMountPointWithVolume", "umount", err)
	}
	DeleteMountPoint(mntURL)
}
