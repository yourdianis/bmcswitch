package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServerConfig 服务器配置结构
type ServerConfig struct {
	InternalIP string
	BMCIP      string
	Username   string
	Password   string
}

// Config 配置文件结构
type Config struct {
	Servers []ServerConfig
}

var config Config

// PowerRequest 电源操作请求体
type PowerRequest struct {
	IP string `json:"ip"`
}

// loadConfig 加载配置文件
func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 按管道符分割
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			return fmt.Errorf("配置文件第%d行格式错误，应为: 带内IP | 带外IP | 用户名 | 密码", lineNum)
		}

		// 去除每个字段的前后空格
		internalIP := strings.TrimSpace(parts[0])
		bmcIP := strings.TrimSpace(parts[1])
		username := strings.TrimSpace(parts[2])
		password := strings.TrimSpace(parts[3])

		if internalIP == "" || bmcIP == "" || username == "" || password == "" {
			return fmt.Errorf("配置文件第%d行存在空字段", lineNum)
		}

		config.Servers = append(config.Servers, ServerConfig{
			InternalIP: internalIP,
			BMCIP:      bmcIP,
			Username:   username,
			Password:   password,
		})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取配置文件时出错: %v", err)
	}

	return nil
}

// findServerByInternalIP 根据带内IP查找服务器配置
func findServerByInternalIP(internalIP string) (*ServerConfig, error) {
	for _, server := range config.Servers {
		if server.InternalIP == internalIP {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("未找到带内IP: %s 对应的配置", internalIP)
}

// executeIPMICommand 执行IPMI命令
func executeIPMICommand(server *ServerConfig, action string) error {
	var cmd *exec.Cmd

	switch action {
	case "on":
		cmd = exec.Command("ipmitool", "-I", "lanplus", "-H", server.BMCIP,
			"-U", server.Username, "-P", server.Password, "power", "on")
	case "off":
		cmd = exec.Command("ipmitool", "-I", "lanplus", "-H", server.BMCIP,
			"-U", server.Username, "-P", server.Password, "power", "off")
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行IPMI命令失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// getIPMIStatus 获取电源状态，返回标准化状态(on/off/unknown)和原始输出
func getIPMIStatus(server *ServerConfig) (string, string, error) {
	cmd := exec.Command("ipmitool", "-I", "lanplus", "-H", server.BMCIP,
		"-U", server.Username, "-P", server.Password, "power", "status")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("执行IPMI状态命令失败: %v, 输出: %s", err, string(output))
	}

	raw := strings.TrimSpace(string(output))
	lower := strings.ToLower(raw)
	status := "unknown"
	if strings.Contains(lower, "is on") {
		status = "on"
	} else if strings.Contains(lower, "is off") {
		status = "off"
	}

	return status, raw, nil
}

// powerOnHandler 开机接口
func powerOnHandler(c *gin.Context) {
	var req PowerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求体格式错误，需要 JSON: {\"ip\": \"带内IP\"}",
		})
		return
	}

	internalIP := strings.TrimSpace(req.IP)
	if internalIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: ip"})
		return
	}

	server, err := findServerByInternalIP(internalIP)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := executeIPMICommand(server, "on"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "开机命令执行成功",
		"internal_ip": internalIP,
		"bmc_ip":     server.BMCIP,
	})
}

// powerOffHandler 关机接口
func powerOffHandler(c *gin.Context) {
	var req PowerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求体格式错误，需要 JSON: {\"ip\": \"带内IP\"}",
		})
		return
	}

	internalIP := strings.TrimSpace(req.IP)
	if internalIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: ip"})
		return
	}

	server, err := findServerByInternalIP(internalIP)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := executeIPMICommand(server, "off"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "关机命令执行成功",
		"internal_ip": internalIP,
		"bmc_ip":     server.BMCIP,
	})
}

// powerStatusHandler 查看电源状态接口
func powerStatusHandler(c *gin.Context) {
	var req PowerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求体格式错误，需要 JSON: {\"ip\": \"带内IP\"}",
		})
		return
	}

	internalIP := strings.TrimSpace(req.IP)
	if internalIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: ip"})
		return
	}

	server, err := findServerByInternalIP(internalIP)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status, raw, err := getIPMIStatus(server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"internal_ip": internalIP,
		"bmc_ip":      server.BMCIP,
		"status":      status,   // on / off / unknown
		"raw_output":  raw,      // ipmitool 原始输出
	})
}

func main() {
	// 加载配置文件
	if err := loadConfig("config.txt"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建Gin路由
	r := gin.Default()

	// 定义路由
	r.POST("/on", powerOnHandler)
	r.POST("/off", powerOffHandler)
	r.POST("/status", powerStatusHandler)

	// 获取端口，默认8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	fmt.Printf("IPMITool服务启动在 :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

