package device

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"wol-pro/internal/wol"
)

// Device представляет устройство для WOL
type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MacAddr   string `json:"mac_address"`
	IPAddr    string `json:"ip_address,omitempty"`
	CreatedAt string `json:"created_at"`
}

// Manager управляет списком устройств
type Manager struct {
	devices  map[string]*Device
	mu       sync.RWMutex
	filepath string
}

// NewManager создает новый менеджер устройств
func NewManager() *Manager {
	m := &Manager{
		devices: make(map[string]*Device),
	}
	m.filepath = m.getConfigPath()
	m.Load()
	return m
}

// getConfigPath возвращает путь к файлу конфигурации
func (m *Manager) getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "wol-pro")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "devices.json")
}

// Add добавляет новое устройство
func (m *Manager) Add(name, macAddr, ipAddr string) (string, error) {
	// Валидация
	if name == "" {
		return "", fmt.Errorf("название устройства не может быть пустым")
	}

	if err := wol.ValidateMAC(macAddr); err != nil {
		return "", fmt.Errorf("некорректный MAC адрес: %w", err)
	}

	// Нормализуем MAC адрес
	normalizedMAC, _ := wol.NormalizeMAC(macAddr)

	// Генерируем уникальный ID на основе времени и порядкового номера
	id := fmt.Sprintf("dev_%d", time.Now().UnixNano())

	device := &Device{
		ID:        id,
		Name:      name,
		MacAddr:   normalizedMAC,
		IPAddr:    ipAddr,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	m.mu.Lock()
	m.devices[id] = device
	m.save()
	m.mu.Unlock()

	return id, nil
}

// Get получает устройство по ID
func (m *Manager) Get(id string) *Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.devices[id]
}

// GetAll возвращает все устройства
func (m *Manager) GetAll() []*Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*Device, 0, len(m.devices))
	for _, dev := range m.devices {
		devices = append(devices, dev)
	}
	return devices
}

// Update обновляет устройство
func (m *Manager) Update(id, name, macAddr, ipAddr string) error {
	if err := wol.ValidateMAC(macAddr); err != nil {
		return fmt.Errorf("некорректный MAC адрес: %w", err)
	}

	normalizedMAC, _ := wol.NormalizeMAC(macAddr)

	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[id]
	if !exists {
		return fmt.Errorf("устройство не найдено")
	}

	device.Name = name
	device.MacAddr = normalizedMAC
	device.IPAddr = ipAddr

	m.save()
	return nil
}

// Delete удаляет устройство
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[id]; !exists {
		return fmt.Errorf("устройство не найдено")
	}

	delete(m.devices, id)
	m.save()
	return nil
}

// Save сохраняет устройства в JSON файл (потокобезопасный, вызывается извне)
func (m *Manager) Save() error {
	m.mu.RLock()
	err := m.saveInternal()
	m.mu.RUnlock()
	return err
}

// save сохраняет устройства в JSON файл (вызывается под мьютексом)
func (m *Manager) save() error {
	return m.saveInternal()
}

// saveInternal внутренняя реализация сохранения (мьютекс должен быть захвачен вызывающей стороной)
func (m *Manager) saveInternal() error {
	devices := make([]*Device, 0, len(m.devices))
	for _, dev := range m.devices {
		devices = append(devices, dev)
	}

	log.Printf("💾 Сохранение %d устройств в %s", len(devices), m.filepath)

	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		log.Printf("❌ Ошибка сериализации: %v", err)
		return err
	}

	err = os.WriteFile(m.filepath, data, 0644)
	if err != nil {
		log.Printf("❌ Ошибка записи файла: %v", err)
		return err
	}

	log.Printf("✅ Сохранено успешно")
	return nil
}

// Load загружает устройства из JSON файла
func (m *Manager) Load() error {
	log.Printf("📖 Загрузка устройств из %s", m.filepath)

	data, err := os.ReadFile(m.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("ℹ️  Файл конфигурации еще не существует")
			return nil // Файл еще не создан, это нормально
		}
		log.Printf("❌ Ошибка чтения файла: %v", err)
		return err
	}

	var devices []*Device
	if err := json.Unmarshal(data, &devices); err != nil {
		log.Printf("❌ Ошибка десериализации: %v", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, dev := range devices {
		m.devices[dev.ID] = dev
		log.Printf("  ✓ Загружено: %s (%s)", dev.Name, dev.ID)
	}

	log.Printf("✅ Загружено %d устройств", len(devices))
	return nil
}
