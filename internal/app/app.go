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

// Result представляет единый ответ API для фронтенда.
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
	MAC     string `json:"mac,omitempty"`
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
func (a *AppInstance) AddDevice(name, macAddr, ipAddr string) Result {
	log.Printf("➕ Добавление устройства: %s (MAC: %s)", name, macAddr)
	id, err := a.manager.Add(name, macAddr, ipAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: %v", err)
		log.Printf("❌ %s", errMsg)
		return Result{
			Success: false,
			Message: errMsg,
		}
	}
	msg := fmt.Sprintf("✅ Устройство '%s' добавлено", name)
	log.Print(msg)
	return Result{
		Success: true,
		ID:      id,
		Message: msg,
	}
}

// UpdateDevice обновляет устройство
func (a *AppInstance) UpdateDevice(id, name, macAddr, ipAddr string) Result {
	log.Printf("✏️  UpdateDevice вызван для ID: %s", id)
	log.Printf("  Новые данные: %s (MAC: %s, IP: %s)", name, macAddr, ipAddr)
	err := a.manager.Update(id, name, macAddr, ipAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: %v", err)
		log.Printf("❌ %s", errMsg)
		return Result{
			Success: false,
			Message: errMsg,
		}
	}
	msg := fmt.Sprintf("✅ Устройство '%s' обновлено", name)
	log.Print(msg)
	return Result{
		Success: true,
		Message: msg,
	}
}

// DeleteDevice удаляет устройство
func (a *AppInstance) DeleteDevice(id string) Result {
	log.Printf("🗑️  DeleteDevice вызван для ID: %s", id)
	err := a.manager.Delete(id)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: %v", err)
		log.Printf("❌ %s", errMsg)
		return Result{
			Success: false,
			Message: errMsg,
		}
	}
	msg := "✅ Устройство удалено"
	log.Print(msg)
	return Result{
		Success: true,
		Message: msg,
	}
}

// SendWOL отправляет WOL пакет
func (a *AppInstance) SendWOL(id string) Result {
	log.Printf("📤 SendWOL вызван для ID: %s", id)
	dev := a.manager.Get(id)
	if dev == nil {
		return Result{
			Success: false,
			Message: "Устройство не найдено",
		}
	}

	success, msg := wol.SendMagicPacket(dev.MacAddr, dev.IPAddr)
	if success {
		log.Printf("✅ WOL пакет отправлен на %s", dev.Name)
		return Result{
			Success: true,
			Message: msg,
		}
	}
	log.Printf("❌ Ошибка отправки WOL: %s", msg)
	return Result{
		Success: false,
		Message: msg,
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
func (a *AppInstance) ScanMACByIP(ipAddr string) Result {
	log.Printf("🔍 Сканирование MAC для IP: %s", ipAddr)
	mac, err := network.ScanMACByIP(ipAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка сканирования: %v", err)
		log.Printf("❌ %s", errMsg)
		return Result{
			Success: false,
			Message: errMsg,
		}
	}
	successMsg := fmt.Sprintf("✅ MAC адрес найден: %s", mac)
	log.Print(successMsg)
	return Result{
		Success: true,
		MAC:     mac,
		Message: successMsg,
	}
}
