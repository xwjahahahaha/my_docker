package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xwj/mydocker/log"
	"xwj/mydocker/utils"
)

// NewWorkSpace
// @Description: 创建新的文件工作空间
// @param rootURL
// @param mntURL
// @param volume 是否使用数据卷
func NewWorkSpace(rootURL, ImageTarPath, mntURL, volume, cId string) {
	// 验证tar包路径的合法性并返回镜像包名称
	imageName := VerifyImageTar(ImageTarPath)
	if imageName == "" {
		return
	}
	CreateReadOnlyLayer(rootURL, ImageTarPath, imageName)      // 创建init只读层
	CreateWriteLayer(rootURL, cId)                    			// 创建读写层
	CreateMountPoint(rootURL, imageName, mntURL, cId) 			// 创建mnt文件夹并挂载
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
	if has, err := utils.DirOrFileExist(parentUrl); err == nil && !has {
		// 当宿主机没有此文件时，创建文件夹
		if err := os.Mkdir(parentUrl, 0777); err != nil {
			log.LogErrorFrom("MountVolume", "Mkdir", err)
			return
		}
	}
	// 2. 在容器目录中创建挂载点目录
	if has, err := utils.DirOrFileExist(containerUrl); err == nil && has {
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
// @param ImageTarPath
func CreateReadOnlyLayer(rootURL, ImageTarPath, imageName string) {
	imageDir := filepath.Join(rootURL, "diff", imageName)
	if has, err := utils.DirOrFileExist(imageDir); err == nil && !has {
		// 如果不存在就循环的创建文件夹
		if err := os.MkdirAll(imageDir, 0777); err != nil {
			log.LogErrorFrom("createReadOnlyLayer", "Mkdir", err)
		}
	}
	if _, err := exec.Command("tar", "-xvf", ImageTarPath, "-C", imageDir).CombinedOutput(); err != nil {
		log.LogErrorFrom("createReadOnlyLayer", "tar", err)
	}
}

// VerifyImageTar
// @Description:  验证tar包路径的合法性并返回镜像包名称
// @param ImageTarPath
// @return string
func VerifyImageTar(ImageTarPath string) string {
	if has, err := utils.DirOrFileExist(ImageTarPath); err != nil {
		log.LogErrorFrom("VerifyImageTar", "dirOrFileExist", err)
		return ""
	}else if err == nil && !has {
		log.LogErrorFrom("VerifyImageTar", "dirOrFileExist", fmt.Errorf(" Not found this image tar!"))
		return ""
	}
	paths := strings.Split(ImageTarPath, "/")
	tarFileName := paths[len(paths)-1]
	if !strings.HasSuffix(tarFileName, "tar") {
		log.LogErrorFrom("VerifyImageTar", "HasPrefix", fmt.Errorf(" not a tar file!"))
	}
	return strings.Split(tarFileName, ".")[0]
}

// CreateWriteLayer
// @Description: 创建读写层
// @param rootURL
func CreateWriteLayer(rootURL, cId string) {
	writeURL := filepath.Join(rootURL, "diff", cId + "_writeLayer")
	if has, err := utils.DirOrFileExist(writeURL); err == nil && has {
		log.Log.Info("Write layer dir already exist. Delete and create new one.")
		// 如果存在则先删除掉之前的
		DeleteWriteLayer(rootURL, cId)
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
func CreateMountPoint(rootURL, imageName, mntURL, cId string) {
	if has, err := utils.DirOrFileExist(mntURL); err == nil && has {
		log.Log.Info("mnt dir already exist. Delete and create new one.")
		DeleteMountPoint(mntURL)
	}
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		log.LogErrorFrom("CreateMountPoint", "Mkdir", err)
	}
	// 将读写层目录与镜像只读层目录mount到mnt目录下
	writerPath := filepath.Join(rootURL, "diff", cId + "_writeLayer")
	imageDir := filepath.Join(rootURL, "diff", imageName)
	dirs := "dirs=" + writerPath + ":" + imageDir
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "mnt_" + cId[:4], mntURL)
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
func DeleteWorkSpace(rootURL, mntURL, volume, cId string) {
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
	DeleteWriteLayer(rootURL, cId)
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
func DeleteWriteLayer(rootURL, cId string) {
	writerPath := filepath.Join(rootURL, "diff", cId + "_writeLayer")
	if err := os.RemoveAll(writerPath); err != nil {
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
