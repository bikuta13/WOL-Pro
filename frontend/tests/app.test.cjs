const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

function deferred() {
    let resolve, reject;
    const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
    return { promise, resolve, reject };
}

function harness(api = {}) {
    const elements = new Map();
    const makeElement = () => {
        const classes = new Set();
        return {
            value: '', textContent: '', disabled: false, style: {}, dataset: {}, children: [],
            classList: { add: name => classes.add(name), remove: name => classes.delete(name), contains: name => classes.has(name) },
            appendChild(child) { this.children.push(child); },
            addEventListener() {}, focus() {},
        };
    };
    const element = id => {
        if (!elements.has(id)) elements.set(id, makeElement());
        return elements.get(id);
    };
    const context = vm.createContext({
        console: { error() {} }, setTimeout() {},
        document: { readyState: 'loading', addEventListener() {}, getElementById: element, createElement: makeElement, querySelectorAll: () => [] },
        window: { go: { app: { AppInstance: { GetDevices: async () => [], ...api } } } },
    });
    const source = fs.readFileSync(path.join(__dirname, '../dist/app.js'), 'utf8');
    vm.runInContext(source + '\nWOLApp.prototype.init = function() {}; globalThis.TestApp = WOLApp;', context);
    return { app: new context.TestApp(), element };
}

function fillDevice(app, element) {
    app.onAddClick();
    element('deviceName').value = 'PC';
    element('deviceMac').value = '00:11:22:33:44:55';
}

test('double Save sends one request and locks the form until completion', async () => {
    const pending = deferred();
    let calls = 0;
    const { app, element } = harness({ AddDevice: () => { calls++; return pending.promise; } });
    fillDevice(app, element);
    const first = app.onSaveDevice();
    await app.onSaveDevice();
    app.closeModal();
    assert.equal(calls, 1);
    assert.equal(element('btnSaveDevice').disabled, true);
    assert.equal(element('deviceModal').classList.contains('open'), true);
    pending.resolve({ success: true, message: 'Saved' });
    await first;
    assert.equal(element('btnSaveDevice').disabled, false);
    assert.equal(element('deviceModal').classList.contains('open'), false);
    await app.onSaveDevice();
    assert.equal(calls, 1, 'closed form must not submit again');
});

test('failed Save keeps data and allows retry', async () => {
    let calls = 0;
    const { app, element } = harness({ AddDevice: async () => { calls++; throw new Error('disk full'); } });
    fillDevice(app, element);
    await app.onSaveDevice();
    assert.equal(element('deviceName').value, 'PC');
    assert.equal(element('deviceModal').classList.contains('open'), true);
    assert.equal(element('btnSaveDevice').disabled, false);
    await app.onSaveDevice();
    assert.equal(calls, 2);
});

test('edit uses the device opened in the form', async () => {
    let updated;
    const { app } = harness({ UpdateDevice: async id => { updated = id; return { success: true }; } });
    app.devices = [{ id: 'one', name: 'PC', mac_address: '00:11:22:33:44:55' }];
    app.selectedDeviceId = 'one';
    app.onEditClick();
    app.selectedDeviceId = 'two';
    await app.onSaveDevice();
    assert.equal(updated, 'one');
});

test('delete uses the confirmed device and ignores a second confirmation', async () => {
    const ids = [];
    const { app } = harness({ DeleteDevice: async id => { ids.push(id); return { success: true }; } });
    app.devices = [{ id: 'one', name: 'PC' }];
    app.selectedDeviceId = 'one';
    await app.onDeleteClick();
    app.selectedDeviceId = 'two';
    const first = app.onConfirmYes();
    await app.onConfirmYes();
    await first;
    assert.deepEqual(ids, ['one']);
});

test('changing IP prevents use of a previous MAC result', async () => {
    const { app, element } = harness({ ScanMACByIP: async () => ({ success: true, mac: '00:11:22:33:44:55' }) });
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    await app.onStartScan();
    element('scanIP').value = '192.168.1.20';
    app.useScanResult();
    assert.equal(element('deviceModal').classList.contains('open'), false);
    assert.equal(app.scanResult, null);
});

test('scan result preserves its IP and MAC as a pair', async () => {
    const { app, element } = harness({ ScanMACByIP: async () => ({ success: true, mac: '00:11:22:33:44:55' }) });
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    await app.onStartScan();
    app.useScanResult();
    assert.equal(element('deviceIp').value, '192.168.1.10');
    assert.equal(element('deviceMac').value, '00:11:22:33:44:55');
    assert.equal(element('deviceModal').classList.contains('open'), true);
});

test('repeated scan events cannot start a duplicate request', async () => {
    const pending = deferred();
    let calls = 0;
    const { app, element } = harness({ ScanMACByIP: () => { calls++; return pending.promise; } });
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    const first = app.onStartScan();
    await app.onStartScan();
    assert.equal(calls, 1);
    pending.resolve({ success: false });
    await first;
    assert.equal(element('btnStartScan').disabled, false);
});

test('late response after closing and reopening scanner is ignored', async () => {
    const pending = deferred();
    const { app, element } = harness({ ScanMACByIP: () => pending.promise });
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    const first = app.onStartScan();
    app.closeScanModal();
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    pending.resolve({ success: true, mac: '00:11:22:33:44:55' });
    await first;
    assert.equal(app.scanResult, null);
    assert.equal(element('scanResult').style.display, 'none');
});

test('old scan cannot overwrite a newer request or unlock its button', async () => {
    const old = deferred(), current = deferred();
    const { app, element } = harness({ ScanMACByIP: ip => ip.endsWith('.10') ? old.promise : current.promise });
    app.openScanModal();
    element('scanIP').value = '192.168.1.10';
    const first = app.onStartScan();
    element('scanIP').value = '192.168.1.20';
    app.invalidateScan();
    const second = app.onStartScan();
    old.resolve({ success: true, mac: '00:11:22:33:44:55' });
    await first;
    assert.equal(element('btnStartScan').disabled, true);
    assert.equal(app.scanResult, null);
    current.resolve({ success: true, mac: 'aa:bb:cc:dd:ee:ff' });
    await second;
    assert.equal(app.scanResult.ip, '192.168.1.20');
    assert.equal(app.scanResult.mac, 'aa:bb:cc:dd:ee:ff');
});

test('selection stays visible after refresh and clears if device disappears', () => {
    const { app, element } = harness();
    app.devices = [{ id: 'one', name: 'PC' }];
    app.selectedDeviceId = 'one';
    app.renderTable();
    assert.equal(element('tableBody').children[0].classList.contains('selected'), true);
    app.devices = [];
    app.renderTable();
    assert.equal(app.selectedDeviceId, null);
});

test('load error is displayed instead of reporting ready', async () => {
    const { app, element } = harness({ GetDevices: async () => { throw new Error('invalid devices.json'); } });
    await app.loadDevices();
    assert.match(element('alert').textContent, /invalid devices.json/);
    assert.equal(element('statusIcon').textContent, '❌');
});
