package app

import (
	"context"
	"fmt"
	"log"

	"wol-pro/internal/device"
	"wol-pro/internal/network"
	"wol-pro/internal/wol"
)

// AppInstance управляет приложением
type AppInstance struct {
	ctx     context.Context
	manager *device.Manager
}

// NewApp создает новое приложение
func NewApp() *AppInstance {
	return &AppInstance{
		manager: device.NewManager(),
	}
}

// Startup вызывается при запуске приложения
func (a *AppInstance) Startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("✅ Приложение запущено")
}

// Shutdown вызывается при завершении приложения
func (a *AppInstance) Shutdown(ctx context.Context) {
	log.Println("👋 Приложение закрыто")
}

// Device представляет устройство для фронтенда
type Device struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	MacAddr string `json:"mac_address"`
	IPAddr  string `json:"ip_address"`
}

// GetDevices возвращает все устройства
func (a *AppInstance) GetDevices() []Device {
	devices := a.manager.GetAll()
	log.Printf("📋 GetDevices вызван, всего устройств: %d", len(devices))
	result := make([]Device, len(devices))
	for i, dev := range devices {
		result[i] = Device{
			ID:      dev.ID,
			Name:    dev.Name,
			MacAddr: dev.MacAddr,
			IPAddr:  dev.IPAddr,
		}
		log.Printf("  - %s (%s)", dev.Name, dev.ID)
	}
	return result
}

// AddDevice добавляет новое устройство
func (a *AppInstance) AddDevice(name, macAddr, ipAddr string) map[string]string {
	log.Printf("➕ Добавление устройства: %s (MAC: %s)", name, macAddr)
	id, err := a.manager.Add(name, macAddr, ipAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: %v", err)
		log.Printf("❌ %s", errMsg)
		return map[string]string{
			"success": "false",
			"message": errMsg,
		}
	}
	msg := fmt.Sprintf("✅ Устройство '%s' добавлено", name)
	log.Printf(msg)
	return map[string]string{
		"success": "true",
		"id":      id,
		"message": msg,
	}
}

// UpdateDevice обновляет устройство
func (a *AppInstance) UpdateDevice(id, name, macAddr, ipAddr string) string {
	log.Printf("✏️  UpdateDevice вызван для ID: %s", id)
	log.Printf("  Новые данные: %s (MAC: %s, IP: %s)", name, macAddr, ipAddr)
	err := a.manager.Update(id, name, macAddr, ipAddr)
	if err != nil {
		log.Printf("❌ Ошибка обновления: %v", err)
		return fmt.Sprintf("Ошибка: %v", err)
	}
	log.Printf("✅ Устройство обновлено")
	return fmt.Sprintf("✅ Устройство '%s' обновлено", name)
}

// DeleteDevice удаляет устройство
func (a *AppInstance) DeleteDevice(id string) string {
	log.Printf("🗑️  DeleteDevice вызван для ID: %s", id)
	err := a.manager.Delete(id)
	if err != nil {
		log.Printf("❌ Ошибка удаления: %v", err)
		return fmt.Sprintf("Ошибка: %v", err)
	}
	log.Printf("✅ Устройство удалено")
	return "✅ Устройство удалено"
}

// SendWOL отправляет WOL пакет
func (a *AppInstance) SendWOL(id string) map[string]string {
	log.Printf("📤 SendWOL вызван для ID: %s", id)
	dev := a.manager.Get(id)
	if dev == nil {
		return map[string]string{
			"success": "false",
			"message": "Устройство не найдено",
		}
	}

	success, msg := wol.SendMagicPacket(dev.MacAddr, dev.IPAddr)
	if success {
		log.Printf("✅ WOL пакет отправлен на %s", dev.Name)
		return map[string]string{
			"success": "true",
			"message": msg,
		}
	}
	log.Printf("❌ Ошибка отправки WOL: %s", msg)
	return map[string]string{
		"success": "false",
		"message": msg,
	}
}

// GetDevice получает одно устройство
func (a *AppInstance) GetDevice(id string) Device {
	dev := a.manager.Get(id)
	if dev == nil {
		return Device{}
	}
	return Device{
		ID:      dev.ID,
		Name:    dev.Name,
		MacAddr: dev.MacAddr,
		IPAddr:  dev.IPAddr,
	}
}

// ScanMACByIP сканирует MAC адрес по IP адресу
func (a *AppInstance) ScanMACByIP(ipAddr string) map[string]string {
	log.Printf("🔍 Сканирование MAC для IP: %s", ipAddr)
	mac, err := network.ScanMACByIP(ipAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка сканирования: %v", err)
		log.Printf("❌ %s", errMsg)
		return map[string]string{
			"success": "false",
			"message": errMsg,
		}
	}
	successMsg := fmt.Sprintf("✅ MAC адрес найден: %s", mac)
	log.Printf(successMsg)
	return map[string]string{
		"success": "true",
		"mac":     mac,
		"message": successMsg,
	}
}
