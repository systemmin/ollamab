# ollamab

**`ollama`** 本地大模型备份工具。 

将 `ollamab.exe` 丢进 **`ollama`** 目录下可以全局使用（已配置环境变量）。

## 备份

`ollamab` `名称:型号（必填）` `指定输出路径，默认输出当前路径（可选）`

```text
备份参数: ollamab 名称:型号（必填） 指定输出路径，默认输出当前路径（可选）
   示例: ollamab deepseek-r1:1.5b 
   示例: ollamab deepseek-r1:1.5b D:/models
   示例: ollamab lrs33/bce-embedding-base_v1:latest
```

## 删除

`ollamab` `名称:型号` `rm`

```text
删除参数: ollamab 名称:型号 rm
   示例: ollamab deepseek-r1:1.5b rm
```

## 支持

Window，Linux

## 打包

```shell
go build -o ollamab.exe main.go
```