/**
 * WOL Pro - Frontend Application
 * Взаимодействие с Go бэкендом через Wails
 */

class WOLApp {
    constructor() {
        this.selectedDeviceId = null;
        this.isEditMode = false;
        this.devices = [];
        this.scannedMAC = null;
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
        }
    }

    renderTable() {
        const tbody = document.getElementById('tableBody');
        tbody.textContent = '';

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

            if (result.success === "true") {
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
        this.isEditMode = false;
        this.openModal('Добавить устройство');
        document.getElementById('deviceName').value = '';
        document.getElementById('deviceMac').value = '';
        document.getElementById('deviceIp').value = '';
        document.getElementById('deviceName').focus();
    }

    onEditClick() {
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
        this.openModal(`Редактировать: ${device.name}`);
        document.getElementById('deviceName').value = device.name;
        document.getElementById('deviceMac').value = device.mac_address;
        document.getElementById('deviceIp').value = device.ip_address || '';
        document.getElementById('deviceName').focus();
    }

    async onDeleteClick() {
        if (!this.selectedDeviceId) {
            this.showError('Выберите устройство');
            return;
        }

        const device = this.devices.find(d => d.id === this.selectedDeviceId);
        if (!device) {
            this.showError('Устройство не найдено');
            return;
        }

        this.showConfirm(`Удалить устройство "${device.name}"?`);
    }

    showConfirm(message) {
        document.getElementById('confirmMessage').textContent = message;
        document.getElementById('confirmModal').classList.add('open');
    }

    closeConfirmModal() {
        document.getElementById('confirmModal').classList.remove('open');
        this._confirmResolve = null;
    }

    onConfirmYes() {
        this.closeConfirmModal();
        this.executeDelete();
    }

    async executeDelete() {
        try {
            this.setStatus('Удаление...', '⏳');
            const result = await window.go.app.AppInstance.DeleteDevice(this.selectedDeviceId);
            if (result.success !== "true") {
                this.showError(result.message);
                this.setStatus('Ошибка удаления', '❌');
                return;
            }

            this.showSuccess(result.message);
            this.selectedDeviceId = null;
            await this.loadDevices();
        } catch (error) {
            console.error('Ошибка удаления:', error);
            this.showError(`Ошибка: ${error}`);
        }
    }

    async onSaveDevice() {
        const name = document.getElementById('deviceName').value.trim();
        const mac = document.getElementById('deviceMac').value.trim();
        const ip = document.getElementById('deviceIp').value.trim();

        if (!name || !mac) {
            this.showError('Заполните все обязательные поля');
            return;
        }

        try {
            this.setStatus('Сохранение...', '⏳');

            if (this.isEditMode) {
                const result = await window.go.app.AppInstance.UpdateDevice(
                    this.selectedDeviceId, name, mac, ip
                );
                if (result.success === "true") {
                    this.showSuccess(result.message);
                } else {
                    this.showError(result.message);
                    this.setStatus('Ошибка обновления', '❌');
                    return;
                }
            } else {
                const result = await window.go.app.AppInstance.AddDevice(name, mac, ip);
                if (result.success === "true") {
                    this.showSuccess(result.message);
                } else {
                    this.showError(result.message);
                    this.setStatus('Ошибка добавления', '❌');
                    return;
                }
            }

            this.closeModal();
            await this.loadDevices();
        } catch (error) {
            console.error('Ошибка сохранения:', error);
            this.showError(`Ошибка: ${error}`);
        }
    }

    openModal(title) {
        document.getElementById('modalTitle').textContent = title;
        document.getElementById('deviceModal').classList.add('open');
    }

    closeModal() {
        document.getElementById('deviceModal').classList.remove('open');
    }

    onScanClick() {
        this.openScanModal();
    }

    openScanModal() {
        document.getElementById('scanIP').value = '';
        document.getElementById('scanResult').style.display = 'none';
        document.getElementById('scanIP').focus();
        document.getElementById('scanModal').classList.add('open');
    }

    closeScanModal() {
        document.getElementById('scanModal').classList.remove('open');
    }

    async onStartScan() {
        const ipAddr = document.getElementById('scanIP').value.trim();

        if (!ipAddr) {
            this.showError('Введите IP адрес');
            return;
        }

        try {
            document.getElementById('btnStartScan').disabled = true;
            this.setStatus('Сканирование...', '⏳');

            const result = await window.go.app.AppInstance.ScanMACByIP(ipAddr);

            if (result.success === "true") {
                document.getElementById('scanResultMAC').textContent = result.mac;
                document.getElementById('scanResult').style.display = 'block';
                this.showSuccess(result.message);
                this.setStatus(result.message, '✅');
                this.scannedMAC = result.mac;
            } else {
                this.showError(result.message);
                document.getElementById('scanResult').style.display = 'none';
                this.setStatus('Сканирование не удалось', '❌');
            }
        } catch (error) {
            console.error('Ошибка сканирования:', error);
            this.showError(`Ошибка: ${error}`);
            this.setStatus('Ошибка сканирования', '❌');
        } finally {
            document.getElementById('btnStartScan').disabled = false;
        }
    }

    useScanResult() {
        if (this.scannedMAC) {
            this.closeScanModal();
            // Открываем модал добавления с заполненными данными из сканера
            this.isEditMode = false;
            this.openModal('Добавить устройство');
            document.getElementById('deviceName').value = '';
            document.getElementById('deviceMac').value = this.scannedMAC;
            document.getElementById('deviceIp').value = document.getElementById('scanIP').value.trim();
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
