# IPMITool 服务器管理工具

基于Go和Gin框架的IPMI服务器电源管理工具，支持通过带内IP进行服务器的开机和关机操作。

## 功能特性

- 通过带内IP查找对应的带外IP（BMC IP）
- 支持服务器开机（on）和关机（off）操作
- 配置文件管理IP映射关系
- RESTful API接口

## 配置文件格式

配置文件 `config.txt` 格式如下，使用管道符（|）分隔，每行一条记录：

**字段顺序（从左到右的第几列）：**

1. **第1列：带内IP**（in-band IP）
2. **第2列：带外IP**（BMC / iLO / iDRAC 等 out-of-band IP）
3. **第3列：用户名**
4. **第4列：密码**

也可以记成一行格式：  
`带内IP | 带外IP | 用户名 | 密码`

### 简单示例

```
192.168.1.100 | 192.168.1.200 | admin | password123
192.168.1.101 | 192.168.1.201 | admin | password456
```

### 详细说明

- 每行一条记录，**第1列一定是带内IP，第2列一定是带外IP**
- 使用管道符（`|`）分隔各字段
- 支持空行和以 `#` 开头的注释行
- 字段前后空格会自动去除

完整示例：
```
# 这是注释行
192.168.1.100 | 192.168.1.200 | admin | password123
192.168.1.101 | 192.168.1.201 | admin | password456
```

## 安装依赖

```bash
go mod download
```

## 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 开机接口

- **请求方式**: POST  
- **路径**: `/on`  
- **请求体(JSON)**:

```json
{
  "ip": "10.222.5.41"
}
```

- **示例**:

```bash
curl -X POST "http://localhost:8081/on" ^
  -H "Content-Type: application/json" ^
  -d "{\"ip\":\"10.222.5.41\"}"
```

- **成功响应**:

```json
{
  "message": "开机命令执行成功",
  "internal_ip": "10.222.5.41",
  "bmc_ip": "172.51.2.153"
}
```

### 关机接口

- **请求方式**: POST  
- **路径**: `/off`  
- **请求体(JSON)**:

```json
{
  "ip": "10.222.5.41"
}
```

- **示例**:

```bash
curl -X POST "http://localhost:8081/off" ^
  -H "Content-Type: application/json" ^
  -d "{\"ip\":\"10.222.5.41\"}"
```

- **成功响应**:

```json
{
  "message": "关机命令执行成功",
  "internal_ip": "10.222.5.41",
  "bmc_ip": "172.51.2.153"
}
```

### 查看电源状态接口

- **请求方式**: POST  
- **路径**: `/status`  
- **请求体(JSON)**:

```json
{
  "ip": "10.222.5.41"
}
```

- **示例**:

```bash
curl -X POST "http://localhost:8081/status" ^
  -H "Content-Type: application/json" ^
  -d "{\"ip\":\"10.222.5.41\"}"
```

- **成功响应**（示例）:

```json
{
  "internal_ip": "10.222.5.41",
  "bmc_ip": "172.51.2.153",
  "status": "on",
  "raw_output": "Chassis Power is on"
}
```

### 错误响应

- 带内 IP 不存在时:

```json
{
  "error": "未找到带内IP: 10.222.5.99 对应的配置"
}
```

- 请求体缺少 `ip` 字段或格式错误时:

```json
{
  "error": "请求体格式错误，需要 JSON: {\"ip\": \"带内IP\"}"
}
```

## 前置要求

- 系统需要安装 `ipmitool` 工具
- 确保网络可以访问BMC IP地址
- 确保BMC用户名和密码正确

## 注意事项

1. 请妥善保管 `config.txt` 文件，避免泄露BMC密码
2. 建议在生产环境中使用环境变量或加密存储密码
3. 确保ipmitool工具已正确安装并配置
4. 配置文件格式简单，方便批量导入和查找

