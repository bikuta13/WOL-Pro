package device

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	loadErr  error
}

// NewManager создает новый менеджер устройств
func NewManager() *Manager {
	m := &Manager{
		devices: make(map[string]*Device),
	}
	var err error
	m.filepath, err = configPath()
	if err != nil {
		m.loadErr = err
	} else if err := m.Load(); err != nil {
		log.Printf("Не удалось загрузить устройства: %v", err)
	}
	return m
}

// configPath сохраняет совместимость с расположением данных предыдущих версий.
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("не удалось определить домашнюю папку: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "wol-pro")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("не удалось создать папку устройств: %w", err)
	}
	return filepath.Join(configDir, "devices.json"), nil
}

// Err сообщает об ошибке загрузки, которая запрещает перезапись хранилища.
func (m *Manager) Err() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loadErr
}

// Add добавляет новое устройство
func (m *Manager) Add(name, macAddr, ipAddr string) (string, error) {
	name = strings.TrimSpace(name)
	ipAddr = strings.TrimSpace(ipAddr)
	// Валидация
	if name == "" {
		return "", fmt.Errorf("название устройства не может быть пустым")
	}

	if err := wol.ValidateMAC(macAddr); err != nil {
		return "", fmt.Errorf("некорректный MAC адрес: %w", err)
	}

	if err := validateIPv4(ipAddr); err != nil {
		return "", err
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
	defer m.mu.Unlock()

	m.devices[id] = device
	if err := m.save(); err != nil {
		delete(m.devices, id)
		return "", fmt.Errorf("не удалось сохранить устройство: %w", err)
	}

	return id, nil
}

// Get получает устройство по ID
func (m *Manager) Get(id string) *Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dev := m.devices[id]
	if dev == nil {
		return nil
	}
	copy := *dev
	return &copy
}

// GetAll возвращает все устройства
func (m *Manager) GetAll() []*Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*Device, 0, len(m.devices))
	for _, dev := range m.devices {
		copy := *dev
		devices = append(devices, &copy)
	}
	sort.SliceStable(devices, func(i, j int) bool {
		left := strings.ToLower(devices[i].Name)
		right := strings.ToLower(devices[j].Name)
		if left == right {
			if devices[i].CreatedAt == devices[j].CreatedAt {
				return devices[i].ID < devices[j].ID
			}
			return devices[i].CreatedAt < devices[j].CreatedAt
		}
		return left < right
	})
	return devices
}

// Update обновляет устройство
func (m *Manager) Update(id, name, macAddr, ipAddr string) error {
	name = strings.TrimSpace(name)
	ipAddr = strings.TrimSpace(ipAddr)
	if name == "" {
		return fmt.Errorf("название устройства не может быть пустым")
	}

	if err := wol.ValidateMAC(macAddr); err != nil {
		return fmt.Errorf("некорректный MAC адрес: %w", err)
	}

	if err := validateIPv4(ipAddr); err != nil {
		return err
	}

	normalizedMAC, _ := wol.NormalizeMAC(macAddr)

	m.mu.Lock()
	defer m.mu.Unlock()

	device, exists := m.devices[id]
	if !exists {
		return fmt.Errorf("устройство не найдено")
	}

	oldName := device.Name
	oldMacAddr := device.MacAddr
	oldIPAddr := device.IPAddr

	device.Name = name
	device.MacAddr = normalizedMAC
	device.IPAddr = ipAddr

	if err := m.save(); err != nil {
		device.Name = oldName
		device.MacAddr = oldMacAddr
		device.IPAddr = oldIPAddr
		return fmt.Errorf("не удалось сохранить изменения: %w", err)
	}

	return nil
}

// Delete удаляет устройство
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.devices[id]; !exists {
		return fmt.Errorf("устройство не найдено")
	}

	device := m.devices[id]
	delete(m.devices, id)
	if err := m.save(); err != nil {
		m.devices[id] = device
		return fmt.Errorf("не удалось сохранить удаление: %w", err)
	}
	return nil
}

// Save сохраняет устройства в JSON файл (потокобезопасный, вызывается извне)
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveInternal()
}

// save сохраняет устройства в JSON файл (вызывается под мьютексом)
func (m *Manager) save() error {
	return m.saveInternal()
}

// saveInternal внутренняя реализация сохранения (мьютекс должен быть захвачен вызывающей стороной)
func (m *Manager) saveInternal() error {
	if m.loadErr != nil {
		return fmt.Errorf("сохранение заблокировано из-за ошибки загрузки: %w", m.loadErr)
	}
	devices := make([]*Device, 0, len(m.devices))
	for _, dev := range m.devices {
		devices = append(devices, dev)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })

	log.Printf("💾 Сохранение %d устройств в %s", len(devices), m.filepath)

	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		log.Printf("❌ Ошибка сериализации: %v", err)
		return err
	}

	// Сохраняем предыдущую корректную версию до замены основного файла.
	previous, err := os.ReadFile(m.filepath)
	if err == nil {
		if _, err := decodeDevices(previous); err != nil {
			return fmt.Errorf("файл устройств повреждён; исходный файл сохранён: %w", err)
		}
		if err := writeFileSafely(m.filepath+".bak", previous); err != nil {
			return fmt.Errorf("не удалось создать резервную копию: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("не удалось прочитать предыдущую версию: %w", err)
	}
	if err := writeFileSafely(m.filepath, data); err != nil {
		return fmt.Errorf("не удалось записать устройства: %w", err)
	}

	log.Printf("✅ Сохранено успешно")
	return nil
}

// Load загружает устройства из JSON файла
func (m *Manager) Load() (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer func() { m.loadErr = err }()
	log.Printf("📖 Загрузка устройств из %s", m.filepath)

	data, err := os.ReadFile(m.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			m.devices = make(map[string]*Device)
			log.Printf("ℹ️  Файл конфигурации еще не существует")
			return nil // Файл еще не создан, это нормально
		}
		log.Printf("❌ Ошибка чтения файла: %v", err)
		return err
	}

	devices, err := decodeDevices(data)
	if err != nil {
		return fmt.Errorf("не удалось загрузить %s; исходный файл сохранён: %w", m.filepath, err)
	}
	m.devices = devices
	log.Printf("✅ Загружено %d устройств", len(devices))
	return nil
}

func decodeDevices(data []byte) (map[string]*Device, error) {
	var records []*Device
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	if records == nil {
		return nil, fmt.Errorf("ожидался JSON-массив устройств")
	}
	devices := make(map[string]*Device, len(records))
	for i, dev := range records {
		if dev == nil || strings.TrimSpace(dev.ID) == "" || strings.TrimSpace(dev.Name) == "" {
			return nil, fmt.Errorf("запись %d: отсутствуют устройство, ID или название", i+1)
		}
		if _, exists := devices[dev.ID]; exists {
			return nil, fmt.Errorf("повторяющийся ID устройства: %s", dev.ID)
		}
		mac, err := wol.NormalizeMAC(dev.MacAddr)
		if err != nil {
			return nil, fmt.Errorf("устройство %s: %w", dev.ID, err)
		}
		dev.Name = strings.TrimSpace(dev.Name)
		dev.MacAddr = mac
		dev.IPAddr = strings.TrimSpace(dev.IPAddr)
		if err := validateIPv4(dev.IPAddr); err != nil {
			return nil, fmt.Errorf("устройство %s: %w", dev.ID, err)
		}
		devices[dev.ID] = dev
	}
	return devices, nil
}

// writeFileSafely не обрезает исходный файл до успешной записи новой версии.
func writeFileSafely(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".devices-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func validateIPv4(ipAddr string) error {
	if ipAddr == "" {
		return nil
	}

	ip := net.ParseIP(ipAddr)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("некорректный IPv4 адрес: %s", ipAddr)
	}

	return nil
}
