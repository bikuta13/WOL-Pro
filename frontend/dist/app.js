/**
 * WOL Pro - Frontend Application
 * Взаимодействие с Go бэкендом через Wails
 */

class WOLApp {
    constructor() {
        this.selectedDeviceId = null;
        this.isEditMode = false;
        this.devices = [];
        this.scanResult = null;
        this.scanGeneration = 0;
        this.activeScan = null;
        this.isSaving = false;
        this.isDeleting = false;
        this.editingDeviceId = null;
        this.pendingDeleteId = null;
        this.init();
    }

    async init() {
        this.attachEventListeners();
        await this.loadDevices();
    }

    attachEventListeners() {
        document.getElementById('btnWake').addEventListener('click', () => this.onWakeClick());
        document.getElementById('btnAdd').addEventListener('click', () => this.onAddClick());
        document.getElementById('btnEdit').addEventListener('click', () => this.onEditClick());
        document.getElementById('btnDelete').addEventListener('click', () => this.onDeleteClick());
        document.getElementById('btnScan').addEventListener('click', () => this.onScanClick());
        document.getElementById('btnSaveDevice').addEventListener('click', () => this.onSaveDevice());
        document.getElementById('btnCancelModal').addEventListener('click', () => this.closeModal());
        document.getElementById('btnStartScan').addEventListener('click', () => this.onStartScan());
        document.getElementById('btnCancelScan').addEventListener('click', () => this.closeScanModal());
        document.getElementById('btnUseScanMAC').addEventListener('click', () => this.useScanResult());
        document.getElementById('btnConfirmYes').addEventListener('click', () => this.onConfirmYes());
        document.getElementById('btnConfirmNo').addEventListener('click', () => this.closeConfirmModal());

        // Закрытие модальных окон по клику вне их
        document.getElementById('deviceModal').addEventListener('click', (e) => {
            if (e.target.id === 'deviceModal') {
                this.closeModal();
            }
        });
        document.getElementById('scanModal').addEventListener('click', (e) => {
            if (e.target.id === 'scanModal') {
                this.closeScanModal();
            }
        });
        document.getElementById('confirmModal').addEventListener('click', (e) => {
            if (e.target.id === 'confirmModal') {
                this.closeConfirmModal();
            }
        });

        // Загрузка устройств при нажатии Enter в формах
        document.getElementById('deviceName').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.onSaveDevice();
        });
        document.getElementById('deviceMac').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.onSaveDevice();
        });
        document.getElementById('deviceIp').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.onSaveDevice();
        });
        document.getElementById('scanIP').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.onStartScan();
        });
        document.getElementById('scanIP').addEventListener('input', () => this.invalidateScan());
    }

    async loadDevices() {
        try {
            this.setStatus('Загрузка устройств...', '⏳');
            const devices = await window.go.app.AppInstance.GetDevices();
            this.devices = devices || [];
            this.renderTable();
            this.setStatus('Готово', '✅');
        } catch (error) {
            console.error('Ошибка загрузки устройств:', error);
            this.showError(`Ошибка загрузки: ${error}`);
            this.setStatus('Ошибка загрузки устройств', '❌');
        }
    }

    renderTable() {
        const tbody = document.getElementById('tableBody');
        tbody.textContent = '';
        if (!this.devices.some(device => device.id === this.selectedDeviceId)) {
            this.selectedDeviceId = null;
        }

        if (this.devices.length === 0) {
            const row = document.createElement('tr');
            row.className = 'empty-table';

            const cell = document.createElement('td');
            cell.colSpan = 3;
            cell.textContent = 'Устройств не найдено. Добавьте первое устройство';

            row.appendChild(cell);
            tbody.appendChild(row);
            return;
        }

        this.devices.forEach((device) => {
            const row = document.createElement('tr');
            row.dataset.deviceId = device.id;
            if (device.id === this.selectedDeviceId) row.classList.add('selected');
            row.addEventListener('click', () => this.selectDevice(device.id, row));

            [device.name, device.mac_address, device.ip_address || '—'].forEach((value) => {
                const cell = document.createElement('td');
                cell.textContent = value;
                row.appendChild(cell);
            });

            tbody.appendChild(row);
        });
    }

    selectDevice(id, row) {
        // Убираем класс selected у всех строк
        document.querySelectorAll('#tableBody tr').forEach(r => r.classList.remove('selected'));
        // Добавляем класс selected выбранной строке
        row.classList.add('selected');
        this.selectedDeviceId = id;
    }

    async onWakeClick() {
        if (!this.selectedDeviceId) {
            this.showError('Выберите устройство');
            return;
        }

        const device = this.devices.find(d => d.id === this.selectedDeviceId);
        if (!device) {
            this.showError('Устройство не найдено');
            return;
        }

        try {
            this.setStatus(`Отправка пакета на ${device.name}...`, '⏳');
            const result = await window.go.app.AppInstance.SendWOL(this.selectedDeviceId);

            if (result.success) {
                this.showSuccess(result.message);
                this.setStatus(result.message, '✅');
            } else {
                this.showError(result.message);
                this.setStatus('Ошибка отправки', '❌');
            }
        } catch (error) {
            console.error('Ошибка отправки WOL:', error);
            this.showError(`Ошибка: ${error}`);
        }
    }

    onAddClick() {
        if (this.isSaving) return;
        this.isEditMode = false;
        this.editingDeviceId = null;
        this.openModal('Добавить устройство');
        document.getElementById('deviceName').value = '';
        document.getElementById('deviceMac').value = '';
        document.getElementById('deviceIp').value = '';
        document.getElementById('deviceName').focus();
    }

    onEditClick() {
        if (this.isSaving) return;
        if (!this.selectedDeviceId) {
            this.showError('Выберите устройство');
            return;
        }

        const device = this.devices.find(d => d.id === this.selectedDeviceId);
        if (!device) {
            this.showError('Устройство не найдено');
            return;
        }

        this.isEditMode = true;
        this.editingDeviceId = device.id;
        this.openModal(`Редактировать: ${device.name}`);
        document.getElementById('deviceName').value = device.name;
        document.getElementById('deviceMac').value = device.mac_address;
        document.getElementById('deviceIp').value = device.ip_address || '';
        document.getElementById('deviceName').focus();
    }

    async onDeleteClick() {
        if (this.isDeleting) return;
        if (!this.selectedDeviceId) {
            this.showError('Выберите устройство');
            return;
        }

        const device = this.devices.find(d => d.id === this.selectedDeviceId);
        if (!device) {
            this.showError('Устройство не найдено');
            return;
        }

        this.pendingDeleteId = device.id;
        this.showConfirm(`Удалить устройство "${device.name}"?`);
    }

    showConfirm(message) {
        document.getElementById('confirmMessage').textContent = message;
        document.getElementById('confirmModal').classList.add('open');
    }

    closeConfirmModal() {
        document.getElementById('confirmModal').classList.remove('open');
        this.pendingDeleteId = null;
    }

    onConfirmYes() {
        const id = this.pendingDeleteId;
        this.closeConfirmModal();
        if (id) return this.executeDelete(id);
    }

    async executeDelete(id) {
        if (this.isDeleting) return;
        this.isDeleting = true;
        try {
            this.setStatus('Удаление...', '⏳');
            const result = await window.go.app.AppInstance.DeleteDevice(id);
            if (!result.success) {
                this.showError(result.message);
                this.setStatus('Ошибка удаления', '❌');
                return;
            }

            this.showSuccess(result.message);
            if (this.selectedDeviceId === id) this.selectedDeviceId = null;
            await this.loadDevices();
        } catch (error) {
            console.error('Ошибка удаления:', error);
            this.showError(`Ошибка: ${error}`);
            this.setStatus('Ошибка удаления', '❌');
        } finally {
            this.isDeleting = false;
        }
    }

    async onSaveDevice() {
        if (this.isSaving || !document.getElementById('deviceModal').classList.contains('open')) return;
        const name = document.getElementById('deviceName').value.trim();
        const mac = document.getElementById('deviceMac').value.trim();
        const ip = document.getElementById('deviceIp').value.trim();

        if (!name || !mac) {
            this.showError('Заполните все обязательные поля');
            return;
        }

        this.isSaving = true;
        this.setDeviceFormBusy(true);
        try {
            this.setStatus('Сохранение...', '⏳');

            if (this.isEditMode) {
                const result = await window.go.app.AppInstance.UpdateDevice(
                    this.editingDeviceId, name, mac, ip
                );
                if (result.success) {
                    this.showSuccess(result.message);
                } else {
                    this.showError(result.message);
                    this.setStatus('Ошибка обновления', '❌');
                    return;
                }
            } else {
                const result = await window.go.app.AppInstance.AddDevice(name, mac, ip);
                if (result.success) {
                    this.showSuccess(result.message);
                } else {
                    this.showError(result.message);
                    this.setStatus('Ошибка добавления', '❌');
                    return;
                }
            }

            this.closeModal(true);
            await this.loadDevices();
        } catch (error) {
            console.error('Ошибка сохранения:', error);
            this.showError(`Ошибка: ${error}`);
            this.setStatus('Ошибка сохранения', '❌');
        } finally {
            this.isSaving = false;
            this.setDeviceFormBusy(false);
        }
    }

    setDeviceFormBusy(busy) {
        ['deviceName', 'deviceMac', 'deviceIp', 'btnSaveDevice', 'btnCancelModal'].forEach(id => {
            document.getElementById(id).disabled = busy;
        });
    }

    openModal(title) {
        document.getElementById('modalTitle').textContent = title;
        document.getElementById('deviceModal').classList.add('open');
    }

    closeModal(force = false) {
        if (this.isSaving && !force) return;
        document.getElementById('deviceModal').classList.remove('open');
    }

    onScanClick() {
        this.openScanModal();
    }

    openScanModal() {
        this.invalidateScan();
        document.getElementById('scanIP').value = '';
        document.getElementById('scanModal').classList.add('open');
        document.getElementById('scanIP').focus();
    }

    closeScanModal() {
        this.invalidateScan();
        document.getElementById('scanModal').classList.remove('open');
    }

    invalidateScan() {
        this.scanGeneration++;
        this.scanResult = null;
        this.activeScan = null;
        document.getElementById('scanResult').style.display = 'none';
        document.getElementById('btnStartScan').disabled = false;
    }

    async onStartScan() {
        if (this.activeScan !== null) return;
        const ipAddr = document.getElementById('scanIP').value.trim();

        if (!ipAddr) {
            this.showError('Введите IP адрес');
            return;
        }

        this.invalidateScan();
        const generation = this.scanGeneration;
        this.activeScan = generation;
        try {
            document.getElementById('btnStartScan').disabled = true;
            this.setStatus('Сканирование...', '⏳');

            const result = await window.go.app.AppInstance.ScanMACByIP(ipAddr);
            if (generation !== this.scanGeneration || document.getElementById('scanIP').value.trim() !== ipAddr) return;

            if (result.success) {
                document.getElementById('scanResultMAC').textContent = result.mac;
                document.getElementById('scanResult').style.display = 'block';
                this.showSuccess(result.message);
                this.setStatus(result.message, '✅');
                this.scanResult = { mac: result.mac, ip: ipAddr };
            } else {
                this.showError(result.message);
                document.getElementById('scanResult').style.display = 'none';
                this.setStatus('Сканирование не удалось', '❌');
            }
        } catch (error) {
            if (generation !== this.scanGeneration) return;
            console.error('Ошибка сканирования:', error);
            this.showError(`Ошибка: ${error}`);
            this.setStatus('Ошибка сканирования', '❌');
        } finally {
            if (this.activeScan === generation) {
                this.activeScan = null;
                document.getElementById('btnStartScan').disabled = false;
            }
        }
    }

    useScanResult() {
        if (this.isSaving) return;
        const result = this.scanResult;
        if (result && result.ip !== document.getElementById('scanIP').value.trim()) {
            this.invalidateScan();
            this.showError('IP адрес изменён. Выполните сканирование ещё раз.');
            return;
        }
        if (result) {
            this.closeScanModal();
            // Открываем модал добавления с заполненными данными из сканера
            this.isEditMode = false;
            this.editingDeviceId = null;
            this.openModal('Добавить устройство');
            document.getElementById('deviceName').value = '';
            document.getElementById('deviceMac').value = result.mac;
            document.getElementById('deviceIp').value = result.ip;
            document.getElementById('deviceName').focus();
        }
    }

    showError(message) {
        this.showAlert(message, 'error');
    }

    showSuccess(message) {
        this.showAlert(message, 'success');
    }

    showAlert(message, type) {
        const alert = document.getElementById('alert');
        alert.textContent = message;
        alert.className = `alert show alert-${type}`;
        setTimeout(() => {
            alert.classList.remove('show');
        }, 5000);
    }

    setStatus(text, icon) {
        document.getElementById('statusText').textContent = text;
        document.getElementById('statusIcon').textContent = icon;
    }
}

// Инициализация приложения
let app;
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        app = new WOLApp();
    });
} else {
    app = new WOLApp();
}
