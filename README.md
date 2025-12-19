# IPMITool 服务器管理工具

基于Go和Gin框架的IPMI服务器电源管理工具，支持通过带内IP进行服务器的电源管理操作。

## 功能特性

- 通过带内IP查找对应的带外IP（BMC IP）
- 支持服务器**开机（on）**、**关机（off）**和**查询电源状态（status）**操作
- 配置文件管理IP映射关系
- RESTful API接口（POST请求）
- 支持Docker容器化部署

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

## 部署方式

### 方式一：Docker 部署（推荐）

#### 使用 docker-compose（最简单）

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 使用 docker 命令

1. **构建镜像**：
```bash
docker build -t ipmitool:latest .
```

2. **运行容器**（使用主机网络模式，确保可以访问BMC IP）：
```bash
docker run -d \
  --name ipmitool \
  --network host \
  -v $(pwd)/config.txt:/app/config.txt \
  ipmitool:latest
```

3. **使用环境变量自定义端口**：
```bash
docker run -d \
  --name ipmitool \
  -p 8081:8081 \
  -e PORT=8081 \
  --network host \
  -v $(pwd)/config.txt:/app/config.txt \
  ipmitool:latest
```

> **注意**：使用 `--network host` 模式可以让容器直接访问主机网络，方便访问 BMC IP 地址。

### 方式二：本地运行

1. **安装依赖**：
```bash
go mod download
```

2. **安装 ipmitool**：
   - Linux: `sudo apt-get install ipmitool` 或 `sudo yum install ipmitool`
   - macOS: `brew install ipmitool`
   - Windows: 下载安装包或使用包管理器

3. **运行服务**：
```bash
go run main.go
```

服务默认在 `http://localhost:8080` 启动，可通过环境变量 `PORT` 自定义端口。

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
# Linux/macOS
curl -X POST "http://localhost:8080/on" \
  -H "Content-Type: application/json" \
  -d '{"ip":"10.222.5.41"}'

# Windows PowerShell
curl -X POST "http://localhost:8080/on" `
  -H "Content-Type: application/json" `
  -d '{\"ip\":\"10.222.5.41\"}'
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
# Linux/macOS
curl -X POST "http://localhost:8080/off" \
  -H "Content-Type: application/json" \
  -d '{"ip":"10.222.5.41"}'

# Windows PowerShell
curl -X POST "http://localhost:8080/off" `
  -H "Content-Type: application/json" `
  -d '{\"ip\":\"10.222.5.41\"}'
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
# Linux/macOS
curl -X POST "http://localhost:8080/status" \
  -H "Content-Type: application/json" \
  -d '{"ip":"10.222.5.41"}'

# Windows PowerShell
curl -X POST "http://localhost:8080/status" `
  -H "Content-Type: application/json" `
  -d '{\"ip\":\"10.222.5.41\"}'
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

## 接口说明

### 三个核心接口

1. **`POST /on`** - 开机接口
2. **`POST /off`** - 关机接口  
3. **`POST /status`** - 查询电源状态接口

所有接口都使用 POST 请求，请求体为 JSON 格式：`{"ip": "带内IP"}`

## 前置要求

- 系统需要安装 `ipmitool` 工具（Docker 镜像已包含）
- 确保网络可以访问BMC IP地址
- 确保BMC用户名和密码正确

## 注意事项

1. 请妥善保管 `config.txt` 文件，避免泄露BMC密码
2. 建议在生产环境中使用环境变量或加密存储密码
3. 使用 Docker 部署时，确保容器网络可以访问 BMC IP 地址
4. 配置文件格式简单，方便批量导入和查找
5. 所有接口都使用 POST 请求，请求体必须包含 `ip` 字段

