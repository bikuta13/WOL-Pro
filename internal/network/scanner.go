package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// ScanMACByIP сканирует MAC адрес по IP адресу
func ScanMACByIP(ipAddr string) (string, error) {
	// Валидируем IP адрес
	if net.ParseIP(ipAddr) == nil {
		return "", fmt.Errorf("некорректный IP адрес: %s", ipAddr)
	}

	// В зависимости от ОС используем разные команды
	switch {
	case isLinux():
		return scanMACLinux(ipAddr)
	case isDarwin():
		return scanMACDarwin(ipAddr)
	case isWindows():
		return scanMACWindows(ipAddr)
	default:
		return "", fmt.Errorf("неподдерживаемая операционная система")
	}
}

// scanMACDarwin сканирует MAC адрес на macOS
func scanMACDarwin(ipAddr string) (string, error) {
	// Используем arp команду на macOS
	cmd := exec.Command("arp", "-n", ipAddr)
	output, err := cmd.Output()
	if err != nil {
		// IP не в ARP таблице - пробуем ping первым чтобы разрешить адрес
		pingCmd := exec.Command("ping", "-c", "1", "-W", "1", ipAddr)
		_ = pingCmd.Run() // Игнорируем результат

		// Пробуем arp снова
		cmd = exec.Command("arp", "-n", ipAddr)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("устройство %s не отвечает на ping и не найдено в ARP таблице. Проверьте IP адрес и что устройство подключено к сети", ipAddr)
		}
	}

	// Парсим вывод: ipAddr at xx:xx:xx:xx:xx:xx on en0 ...
	re := regexp.MustCompile(`([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("MAC адрес не найден для IP %s", ipAddr)
}

// scanMACLinux сканирует MAC адрес на Linux
func scanMACLinux(ipAddr string) (string, error) {
	// Используем arp-scan или arp команду
	cmd := exec.Command("arp", "-n", ipAddr)
	output, err := cmd.Output()
	if err != nil {
		// Пробуем альтернативный способ
		return scanMACLinuxAlt(ipAddr)
	}

	// Парсим вывод: ipAddr ether xx:xx:xx:xx:xx:xx ...
	re := regexp.MustCompile(`([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("MAC адрес не найден для IP %s", ipAddr)
}

// scanMACLinuxAlt альтернативный метод для Linux
func scanMACLinuxAlt(ipAddr string) (string, error) {
	// Проверяем файл /proc/net/arp
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return "", fmt.Errorf("не удается прочитать ARP таблицу: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ipAddr) {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				mac := fields[3]
				if isValidMAC(mac) {
					return mac, nil
				}
			}
		}
	}

	return "", fmt.Errorf("MAC адрес не найден для IP %s", ipAddr)
}

// scanMACWindows сканирует MAC адрес на Windows
func scanMACWindows(ipAddr string) (string, error) {
	// Используем getmac команду Windows
	cmd := exec.Command("arp", "-a", ipAddr)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения arp: %w", err)
	}

	// Парсим MAC адрес из вывода
	re := regexp.MustCompile(`([0-9a-fA-F]{2}-[0-9a-fA-F]{2}-[0-9a-fA-F]{2}-[0-9a-fA-F]{2}-[0-9a-fA-F]{2}-[0-9a-fA-F]{2})`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 0 {
		// Преобразуем формат из XX-XX-XX-XX-XX-XX в XX:XX:XX:XX:XX:XX
		mac := strings.ReplaceAll(matches[0], "-", ":")
		return mac, nil
	}

	return "", fmt.Errorf("MAC адрес не найден для IP %s", ipAddr)
}

// isValidMAC проверяет валидность MAC адреса
func isValidMAC(mac string) bool {
	re := regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}([0-9a-fA-F]{2})$`)
	return re.MatchString(mac)
}

// Helper функции
func isDarwin() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// GetOS возвращает текущую операционную систему
func GetOS() string {
	return runtime.GOOS
}
