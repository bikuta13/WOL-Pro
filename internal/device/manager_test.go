package device

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func testManager(t *testing.T) *Manager {
	t.Helper()
	return &Manager{devices: make(map[string]*Device), filepath: filepath.Join(t.TempDir(), "devices.json")}
}

func addTestDevice(t *testing.T, m *Manager) string {
	t.Helper()
	id, err := m.Add(" Computer ", "AA-BB-CC-DD-EE-FF", " 192.168.1.10 ")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestPersistenceAndBackup(t *testing.T) {
	m := testManager(t)
	id := addTestDevice(t, m)
	before, err := os.ReadFile(m.filepath)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Update(id, "Updated", "00:11:22:33:44:55", ""); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(m.filepath + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, backup) {
		t.Fatal("backup must contain the previous version")
	}
	reloaded := &Manager{filepath: m.filepath}
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Get(id); got == nil || got.Name != "Updated" {
		t.Fatalf("unexpected device: %+v", got)
	}
	if err := reloaded.Delete(id); err != nil {
		t.Fatal(err)
	}
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if len(m.GetAll()) != 0 {
		t.Fatal("Load must replace previous state, including deletions")
	}
}

func TestSnapshotsDoNotShareMutableState(t *testing.T) {
	m := testManager(t)
	id := addTestDevice(t, m)
	one, all := m.Get(id), m.GetAll()
	if one.Name != "Computer" || one.MacAddr != "aa:bb:cc:dd:ee:ff" || one.IPAddr != "192.168.1.10" {
		t.Fatalf("not normalized: %+v", one)
	}
	if err := m.Update(id, "Updated", "00:11:22:33:44:55", ""); err != nil {
		t.Fatal(err)
	}
	if one.Name != "Computer" || all[0].Name != "Computer" {
		t.Fatal("Update changed a returned snapshot")
	}
	one.Name, all[0].Name = "Outside", "Outside"
	if m.Get(id).Name != "Updated" {
		t.Fatal("caller changed manager state")
	}
}

func TestInvalidFileBlocksWritesAndPreservesData(t *testing.T) {
	valid := `{"id":"one","name":"Computer","mac_address":"00:11:22:33:44:55"}`
	for _, content := range []string{"{", "null", "[null]", `[{}]`, "[" + valid + "," + valid + "]", `[{"id":"one","name":"PC","mac_address":"00:11:22:33:44:55:66:77"}]`} {
		t.Run(content, func(t *testing.T) {
			m := testManager(t)
			id := addTestDevice(t, m)
			if err := os.WriteFile(m.filepath, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if err := m.Load(); err == nil {
				t.Fatal("invalid data accepted")
			}
			if m.Err() == nil {
				t.Fatal("load error not exposed")
			}
			if m.Get(id) == nil {
				t.Fatal("failed Load changed existing state")
			}
			if _, err := m.Add("New", "00:11:22:33:44:55", ""); err == nil {
				t.Fatal("Add should fail")
			}
			if err := m.Update(id, "New", "00:11:22:33:44:55", ""); err == nil {
				t.Fatal("Update should fail")
			}
			if err := m.Delete(id); err == nil {
				t.Fatal("Delete should fail")
			}
			if err := m.Save(); err == nil {
				t.Fatal("Save should fail")
			}
			got, err := os.ReadFile(m.filepath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != content {
				t.Fatal("invalid source was overwritten")
			}
			if len(m.GetAll()) != 1 || m.Get(id).Name != "Computer" {
				t.Fatal("failed mutations were not rolled back")
			}
			if err := os.WriteFile(m.filepath, []byte("[]"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := m.Load(); err != nil {
				t.Fatal(err)
			}
			if m.Err() != nil {
				t.Fatal("successful reload did not clear load error")
			}
		})
	}
}

func TestBackupFailureLeavesOriginalAndRollsBack(t *testing.T) {
	m := testManager(t)
	id := addTestDevice(t, m)
	before, err := os.ReadFile(m.filepath)
	if err != nil {
		t.Fatal(err)
	}
	// A directory at the backup path reliably prevents replacement on every OS.
	if err := os.Mkdir(m.filepath+".bak", 0700); err != nil {
		t.Fatal(err)
	}
	if err := m.Update(id, "Changed", "00:11:22:33:44:55", ""); err == nil {
		t.Fatal("expected backup failure")
	}
	after, err := os.ReadFile(m.filepath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || m.Get(id).Name != "Computer" {
		t.Fatal("failed save modified data")
	}
	temps, err := filepath.Glob(filepath.Join(filepath.Dir(m.filepath), ".devices-*.tmp"))
	if err != nil || len(temps) != 0 {
		t.Fatalf("temporary files left behind: %v, %v", temps, err)
	}
}

func TestRejectExternallyCorruptedFileOnSave(t *testing.T) {
	m := testManager(t)
	id := addTestDevice(t, m)
	if err := os.WriteFile(m.filepath, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Delete(id); err == nil {
		t.Fatal("must not overwrite externally corrupted file")
	}
	if m.Get(id) == nil {
		t.Fatal("failed deletion was not rolled back")
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	m := testManager(t)
	id := addTestDevice(t, m)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 12; i++ {
				switch worker {
				case 0:
					if err := m.Update(id, fmt.Sprint(i), "00:11:22:33:44:55", ""); err != nil {
						t.Error(err)
					}
				case 1:
					if err := m.Save(); err != nil {
						t.Error(err)
					}
				default:
					one := m.Get(id)
					_ = one.Name
					for _, dev := range m.GetAll() {
						_ = dev.Name
					}
				}
			}
		}(worker)
	}
	wg.Wait()
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestSortTieUsesID(t *testing.T) {
	m := testManager(t)
	m.devices = map[string]*Device{"b": {ID: "b", Name: "PC"}, "a": {ID: "a", Name: "pc"}}
	for i := 0; i < 20; i++ {
		if got := m.GetAll(); got[0].ID != "a" || got[1].ID != "b" {
			t.Fatal("unstable ordering")
		}
	}
}

func TestInvalidInputDoesNotCreateDevice(t *testing.T) {
	for _, input := range [][3]string{{"  ", "00:11:22:33:44:55", ""}, {"PC", "00:11:22:33:44:55:66:77", ""}, {"PC", "00:11:22:33:44:55", "300.1.2.3"}} {
		m := testManager(t)
		if _, err := m.Add(input[0], input[1], input[2]); err == nil {
			t.Fatalf("accepted %v", input)
		}
		if len(m.GetAll()) != 0 {
			t.Fatal("invalid device was added")
		}
	}
}
