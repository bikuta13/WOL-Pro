package wol

import (
	"fmt"
	"net"
	"strings"
)

// SendMagicPacket отправляет WOL пакет на целевое устройство
func SendMagicPacket(macAddr, ipAddr string) (bool, string) {
	// Парсим MAC адрес
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return false, fmt.Sprintf("Некорректный MAC адрес: %v", err)
	}

	// Собираем magic packet
	packet := createMagicPacket(hwAddr)

	// Формируем список адресов для отправки:
	// 1. Глобальный широковещательный адрес
	// 2. Указанный IP-адрес (unicast)
	// 3. Широковещательный адрес подсети (если удалось распарсить IP)
	targetAddrs := []string{"255.255.255.255:9"}

	if ipAddr != "" && ipAddr != "255.255.255.255" {
		targetAddrs = append(targetAddrs, ipAddr+":9")

		// Пытаемся сформировать subnet broadcast (x.x.x.255)
		parts := strings.Split(ipAddr, ".")
		if len(parts) == 4 {
			subnetBcast := fmt.Sprintf("%s.%s.%s.255:9", parts[0], parts[1], parts[2])
			if subnetBcast != ipAddr+":9" {
				targetAddrs = append(targetAddrs, subnetBcast)
			}
		}
	}

	successCount := 0
	var lastErr error

	// Отправляем пакет на все собранные адреса
	for _, target := range targetAddrs {
		addr, err := net.ResolveUDPAddr("udp", target)
		if err != nil {
			lastErr = err
			continue
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			lastErr = err
			continue
		}

		_, err = conn.Write(packet)
		conn.Close()

		if err == nil {
			successCount++
		} else {
			lastErr = err
		}
	}

	if successCount == 0 {
		return false, fmt.Sprintf("Ошибка отправки пакета: %v", lastErr)
	}

	return true, fmt.Sprintf("WOL пакет успешно отправлен на %s (через %d адресов)", macAddr, successCount)
}

// createMagicPacket создает magic packet для WOL
// Magic packet состоит из 6 байт 0xFF следуемых MAC адресом повторенным 16 раз
func createMagicPacket(hwAddr net.HardwareAddr) []byte {
	packet := make([]byte, 102)
	// Первые 6 байт - 0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// Остаток - MAC адрес повторяется 16 раз
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], hwAddr)
	}
	return packet
}

// ValidateMAC проверяет формат MAC адреса
func ValidateMAC(macStr string) error {
	_, err := net.ParseMAC(macStr)
	return err
}

// NormalizeMAC нормализует MAC адрес к формату XX:XX:XX:XX:XX:XX
func NormalizeMAC(macStr string) (string, error) {
	// Удаляем дефисы и заменяем их на двоеточия
	macStr = strings.ReplaceAll(macStr, "-", ":")
	// Удаляем все пробелы
	macStr = strings.ReplaceAll(macStr, " ", "")

	hwAddr, err := net.ParseMAC(macStr)
	if err != nil {
		return "", err
	}

	return hwAddr.String(), nil
}
