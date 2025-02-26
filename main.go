/**
 * @Time : 2025/2/14 10:14
 * @File : main.go
 * @Software: ollamab
 * @Author : Mr.Fang
 * @Description: 备份 ollama 模型
 */

package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var modelPath = ""

// 初始化获取模型文件路径，优先从系统环境变量获取，其次获取默认路径
func init() {
	models := os.Getenv("OLLAMA_MODELS")
	if len(models) > 0 {
		modelPath = models
	} else {
		// Window
		if runtime.GOOS == "windows" {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Panicln("获取用户主目录失败:", err)
			}
			modelPath = filepath.Join(home, ".ollama", "models")
		} else { // Linux
			modelPath = filepath.Join("/usr/share/ollama/", ".ollama", "models")
		}
	}
	fmt.Println("模型路径:", modelPath)
}

// 将文件或文件夹添加到 zip 中
func addToZip(zipWriter *zip.Writer, filePath, baseFolder string) error {
	// 计算文件的相对路径
	relativePath, err := filepath.Rel(baseFolder, filePath)
	if err != nil {
		return err
	}

	fmt.Println("添加到 ZIP 中:", relativePath)

	// 如果是目录，则创建空目录
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// 如果是目录，创建空目录
	if info.IsDir() {
		_, err := zipWriter.Create(relativePath + "/")
		return err
	}

	// 如果是文件，创建文件条目
	fileInZip, err := zipWriter.Create(relativePath)
	if err != nil {
		return err
	}

	// 打开文件并复制内容到zip中
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 复制文件内容到zip
	_, err = io.Copy(fileInZip, file)
	return err
}

// 模型相关所有文件路径
func blobsPath(modelDataPath, basePath string) []string {
	file, err := os.ReadFile(modelDataPath)
	if err != nil {
		fmt.Println("model data 文件读取错误！", err)
	}
	// 转 map
	var modelData map[string]interface{}
	var blobsPath []string
	err = json.Unmarshal(file, &modelData)
	if err != nil {
		fmt.Println("model data 转换错误！", err)
	}
	// 层数据
	layers := modelData["layers"].([]interface{})
	// 模型详情信息
	layers = append(layers, modelData["config"].(interface{}))

	for _, layer := range layers {
		item := layer.(map[string]interface{})
		digest := item["digest"].(string) // sha256
		digest = strings.ReplaceAll(digest, ":", "-")
		join := filepath.Join(basePath, "blobs", digest)
		// 使用 os.Stat 检查文件是否存在
		fileInfo, _ := os.Stat(join)
		if fileInfo != nil {
			blobsPath = append(blobsPath, join)
		}
	}
	return blobsPath
}

// build 打包 zip
func build(name string, output string, folderPaths []string) {
	fmt.Println("开始打包，耐心等待…………")
	// 创建目标zip文件
	zipFilePath := filepath.Join(output, name)
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		fmt.Println("创建zip文件失败:", err)
		return
	}
	defer zipFile.Close()

	// 创建zip写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 逐个添加文件或目录到zip文件中
	for _, filePath := range folderPaths {
		// 注意：baseFolder 是我们希望在zip文件中根目录
		err := addToZip(zipWriter, filePath, modelPath)
		if err != nil {
			fmt.Println("添加文件到zip失败:", err)
			return
		}
	}
	fmt.Println("zip文件创建成功:", zipFilePath)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("备份参数: ollamab 名称:型号（必填） 指定输出路径，默认输出当前路径（可选）")
		fmt.Println("   示例: ollamab deepseek-r1:1.5b ")
		fmt.Println("   示例: ollamab deepseek-r1:1.5b D:/models")
		fmt.Println("   示例: ollamab lrs33/bce-embedding-base_v1:latest")
		fmt.Println()
		fmt.Println("删除参数: ollamab 名称:型号 rm")
		fmt.Println("   示例: ollamab deepseek-r1:1.5b rm")
		return
	}
	arg := strings.Split(args[0], ":")
	name := arg[0]
	version := arg[1]
	output := "./"
	if len(args) == 2 {
		output = args[1]
	}
	// 配置文件路径
	library := filepath.Join(modelPath, "manifests", "registry.ollama.ai", "library", name, version)
	// 特殊情况，用户自己分享的模型
	contains := strings.Contains(name, "/")
	if contains {
		libs := strings.Split(name, "/")
		library = filepath.Join(modelPath, "manifests", "registry.ollama.ai", libs[0], libs[1], version)
		// 替换 "/" 否则无法创建 zip
		name = strings.ReplaceAll(name, "/", "-")
	}
	folderPaths := blobsPath(library, modelPath)
	// 模型路径
	folderPaths = append(folderPaths, library)

	// 删除模型文件
	if output == "rm" {
		for _, path := range folderPaths {
			fmt.Println(path)
			os.Remove(path)
		}
	} else {
		// 打包
		build(fmt.Sprintf("%s-%s.zip", name, version), output, folderPaths)
	}

}
