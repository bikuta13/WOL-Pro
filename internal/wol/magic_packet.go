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

	// Если IP не указан, используем broadcast адрес
	if ipAddr == "" {
		ipAddr = "255.255.255.255:9"
	} else {
		ipAddr = ipAddr + ":9"
	}

	// Отправляем пакет
	addr, err := net.ResolveUDPAddr("udp", ipAddr)
	if err != nil {
		return false, fmt.Sprintf("Ошибка разрешения адреса: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return false, fmt.Sprintf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return false, fmt.Sprintf("Ошибка отправки пакета: %v", err)
	}

	return true, fmt.Sprintf("WOL пакет успешно отправлен на %s", macAddr)
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
