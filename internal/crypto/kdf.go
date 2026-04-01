package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2迭代次数
	Iterations = 100000
	// 密钥长度（AES-256需要32字节）
	KeyLength = 32
	// 盐值（固定，用于密钥派生）
	Salt = "cloud189-cli-config-encryption-salt-v1"
)

// DeriveKey 从机器特征派生加密密钥
func DeriveKey() ([]byte, error) {
	machineID, err := getMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine ID: %w", err)
	}

	// 使用PBKDF2派生密钥
	key := pbkdf2.Key([]byte(machineID), []byte(Salt), Iterations, KeyLength, sha256.New)

	return key, nil
}

// DeriveKeyWithPassword 使用用户密码派生加密密钥
func DeriveKeyWithPassword(password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// 使用PBKDF2派生密钥
	key := pbkdf2.Key([]byte(password), []byte(Salt), Iterations, KeyLength, sha256.New)

	return key, nil
}

// getMachineID 获取机器唯一标识
func getMachineID() (string, error) {
	var components []string

	// 1. 获取主机名
	hostname, err := os.Hostname()
	if err == nil && hostname != "" {
		components = append(components, hostname)
	}

	// 2. 获取用户名
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username != "" {
		components = append(components, username)
	}

	// 3. 获取操作系统相关的机器ID
	switch runtime.GOOS {
	case "windows":
		machineGUID, err := getWindowsMachineGUID()
		if err == nil && machineGUID != "" {
			components = append(components, machineGUID)
		}
	case "darwin":
		hardwareUUID, err := getMacOSHardwareUUID()
		if err == nil && hardwareUUID != "" {
			components = append(components, hardwareUUID)
		}
	case "linux":
		machineID, err := getLinuxMachineID()
		if err == nil && machineID != "" {
			components = append(components, machineID)
		}
	}

	// 4. 获取第一个网卡的MAC地址（可选）
	macAddr, err := getFirstMACAddress()
	if err == nil && macAddr != "" {
		components = append(components, macAddr)
	}

	if len(components) == 0 {
		return "", fmt.Errorf("failed to get any machine identifiers")
	}

	// 组合所有标识符
	machineID := strings.Join(components, "|")

	// 使用SHA256哈希，确保长度一致
	hash := sha256.Sum256([]byte(machineID))
	return fmt.Sprintf("%x", hash), nil
}

// getWindowsMachineGUID 获取Windows机器GUID
func getWindowsMachineGUID() (string, error) {
	// 尝试从注册表读取MachineGuid
	// HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography\MachineGuid
	data, err := os.ReadFile("/proc/registry/HKEY_LOCAL_MACHINE/SOFTWARE/Microsoft/Cryptography/MachineGuid")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// 备用方案：使用环境变量
	computerName := os.Getenv("COMPUTERNAME")
	if computerName != "" {
		return computerName, nil
	}

	return "", fmt.Errorf("failed to get Windows machine GUID")
}

// getMacOSHardwareUUID 获取macOS硬件UUID
func getMacOSHardwareUUID() (string, error) {
	// 在macOS上，我们可以从IOPlatformUUID获取
	// 但由于需要执行命令，这里使用备选方案
	// 使用主机名和用户名的组合
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if hostname != "" && username != "" {
		return fmt.Sprintf("%s-%s", hostname, username), nil
	}
	return "", fmt.Errorf("failed to get macOS hardware UUID")
}

// getLinuxMachineID 获取Linux机器ID
func getLinuxMachineID() (string, error) {
	// 尝试读取/etc/machine-id
	data, err := os.ReadFile("/etc/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// 备用方案：尝试读取/var/lib/dbus/machine-id
	data, err = os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("failed to get Linux machine ID")
}

// getFirstMACAddress 获取第一个非回环网卡的MAC地址
func getFirstMACAddress() (string, error) {
	// 这个函数在不同平台实现不同
	// 简化版本：返回空，让其他标识符起作用
	return "", fmt.Errorf("not implemented")
}

// HashForComparison 创建用于比较的哈希值
func HashForComparison(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// VerifyHash 验证哈希值
func VerifyHash(data, expectedHash string) bool {
	actualHash := HashForComparison(data)
	return actualHash == expectedHash
}

// 选择哈希函数（用于PBKDF2）
func getHashFunc() func() hash.Hash {
	return sha512.New
}
