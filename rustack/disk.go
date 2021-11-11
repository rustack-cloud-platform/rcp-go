package rustack

import "fmt"

type Disk struct {
	manager        *Manager
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Scsi           string          `json:"scsi"`
	Size           int             `json:"size"`
	Vm             *Vm             `json:"vm"`
	StorageProfile *StorageProfile `json:"storage_profile"`
	Locked         bool            `json:"locked,omitempty"`
}

func NewDisk(name string, size int, storageProfile *StorageProfile) Disk {
	d := Disk{Name: name, Size: size, StorageProfile: storageProfile}
	return d
}

func (m *Manager) GetDisks(extraArgs ...Arguments) (disks []*Disk, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/disk"
	err = m.GetItems(path, args, &disks)
	for i := range disks {
		disks[i].manager = m
	}
	return
}

func (v *Vdc) GetDisks(extraArgs ...Arguments) (disks []*Disk, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	disks, err = v.manager.GetDisks(args)
	return
}

func (m *Manager) GetDisk(id string) (disk *Disk, err error) {
	path := fmt.Sprintf("v1/disk/%s", id)
	err = m.Get(path, Defaults(), &disk)
	if err != nil {
		return
	}
	disk.manager = m
	return
}

func (v *Vm) AttachDisk(disk *Disk) error {
	path := fmt.Sprintf("v1/disk/%s/attach", disk.ID)

	args := &struct {
		Vm string `json:"vm"`
	}{
		Vm: v.ID,
	}

	err := v.manager.Post(path, args, nil)
	if err != nil {
		return err
	}

	return nil
}

func (v *Vm) DetachDisk(disk *Disk) error {
	path := fmt.Sprintf("v1/disk/%s/detach", disk.ID)

	err := v.manager.Post(path, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (d *Disk) update(name string, size int, storageProfileId string) error {
	path := fmt.Sprintf("v1/disk/%s", d.ID)

	args := &struct {
		Name           string `json:"name"`
		Size           int    `json:"size"`
		StorageProfile string `json:"storage_profile"`
	}{
		Name:           name,
		Size:           size,
		StorageProfile: storageProfileId,
	}

	err := d.manager.Put(path, args, d)
	if err != nil {
		return err
	}

	return nil
}

func (d *Disk) Rename(name string) error {
	return d.update(name, d.Size, d.StorageProfile.ID)
}

func (d *Disk) Resize(size int) error {
	return d.update(d.Name, size, d.StorageProfile.ID)
}

func (d *Disk) UpdateStorageProfile(storageProfile StorageProfile) error {
	return d.update(d.Name, d.Size, storageProfile.ID)
}

func (d *Disk) Delete() error {
	path := fmt.Sprintf("v1/disk/%s", d.ID)
	return d.manager.Delete(path, Defaults(), d)
}
