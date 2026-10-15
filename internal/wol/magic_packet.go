package wol

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// SendMagicPacket отправляет WOL пакет на целевое устройство
func SendMagicPacket(macAddr, ipAddr string) (bool, string) {
	// Парсим MAC адрес
	hwAddr, err := parseMAC(macAddr)
	if err != nil {
		return false, fmt.Sprintf("Некорректный MAC адрес: %v", err)
	}

	ipAddr = strings.TrimSpace(ipAddr)
	var ip net.IP
	if ipAddr != "" {
		ip = net.ParseIP(ipAddr).To4()
		if ip == nil {
			return false, fmt.Sprintf("Некорректный IPv4 адрес: %s", ipAddr)
		}
	}
	packet := createMagicPacket(hwAddr)
	targets := packetTargets(ip, localIPv4Networks())

	successCount := 0
	var lastErr error

	// Отправляем пакет на все собранные адреса
	for _, target := range targets {
		var local *net.UDPAddr
		if target.source != nil {
			local = &net.UDPAddr{IP: target.source}
		}
		conn, err := net.DialUDP("udp4", local, &net.UDPAddr{IP: target.destination, Port: 9})
		if err != nil {
			lastErr = err
			continue
		}

		err = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err == nil {
			_, err = conn.Write(packet)
		}
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

	return true, fmt.Sprintf("WOL пакет отправлен для %s (отправок: %d). Включение устройства не подтверждено.", hwAddr, successCount)
}

type packetTarget struct {
	source      net.IP
	destination net.IP
}

// packetTargets вычисляет broadcast по маске локального интерфейса.
// Для удалённых сетей маска неизвестна: адрес x.x.x.255 не угадываем.
func packetTargets(ip net.IP, networks []*net.IPNet) []packetTarget {
	targets := []packetTarget{}
	seen := make(map[string]bool)
	add := func(source, destination net.IP) {
		key := source.String() + "/" + destination.String()
		if !seen[key] {
			seen[key] = true
			targets = append(targets, packetTarget{source, destination})
		}
	}
	if ip != nil && !ip.Equal(net.IPv4bcast) {
		add(nil, ip)
	}
	for _, network := range networks {
		local := network.IP.To4()
		ones, bits := network.Mask.Size()
		if local == nil || bits != 32 || ones >= 31 || (ip != nil && !ip.Equal(net.IPv4bcast) && !network.Contains(ip)) {
			continue
		}
		broadcast := make(net.IP, 4)
		for i := range broadcast {
			broadcast[i] = local[i] | ^network.Mask[i]
		}
		add(local, broadcast)
	}
	add(nil, net.IPv4bcast)
	return targets
}

func localIPv4Networks() []*net.IPNet {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var networks []*net.IPNet
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			if network, ok := address.(*net.IPNet); ok && network.IP.To4() != nil {
				networks = append(networks, network)
			}
		}
	}
	return networks
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
	_, err := parseMAC(macStr)
	return err
}

// NormalizeMAC нормализует MAC адрес к формату XX:XX:XX:XX:XX:XX
func NormalizeMAC(macStr string) (string, error) {
	hwAddr, err := parseMAC(macStr)
	if err != nil {
		return "", err
	}

	return hwAddr.String(), nil
}

func parseMAC(macStr string) (net.HardwareAddr, error) {
	hwAddr, err := net.ParseMAC(strings.TrimSpace(macStr))
	if err != nil {
		return nil, err
	}
	if len(hwAddr) != 6 {
		return nil, fmt.Errorf("для Wake-on-LAN требуется MAC адрес длиной 6 байт")
	}
	return hwAddr, nil
}
