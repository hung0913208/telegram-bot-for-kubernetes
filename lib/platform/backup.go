package platform

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type Backup interface {
	Prepare(uuid string) error
	Backup() error
}

type BackupFullSetup interface {
	Backup

	GetStatus() (map[string]BackupState, error)
	SetStatus(uuid string, status BackupState) error

	SetClient(client kubernetes.Kubernetes)
	SetNamespace(namespace string)
	SetVolumes(volumes []corev1.PersistentVolumeClaim)
	SetCommand(command string)
	SetService(service string)
}

type backupImpl struct {
	client    kubernetes.Kubernetes
	volumes   []corev1.PersistentVolumeClaim
	hook      kubernetes.Hook
	uuid      string
	service   string
	namespace string
	command   string
}

func NewBackup(
	client kubernetes.Kubernetes,
) Backup {
	hook := kubernetes.Hook{
		Header:     DefaultHeaderOfBackupHook,
		Exec:       DefaultExecuterToProvisionBackup,
		Image:      DefaultRsyncImageForBackup,
		MountPoint: DefaultMountPointOfBackupScript,
		PreHook:    DefaultPreHookForBackup,
		PostHook:   DefaultPostHookForaBackup,
	}

	return &backupImpl{
		client: kubernetes.NewFromClient(
			client,
			hook,
		),
		namespace: DefaultNamespaceToLocateBackupJob,
		command:   DefaultCommandToBackupVolumes,
		hook:      hook,
	}
}

func NewCustomBackup(
	client kubernetes.Kubernetes,
	hook kubernetes.Hook,
) Backup {
	return &backupImpl{
		client: kubernetes.NewFromClient(
			client,
			hook,
		),
		hook: hook,
	}
}

func (self *backupImpl) SetNamespace(namespace string) {
	self.namespace = namespace
}

func (self *backupImpl) SetCommand(command string) {
	self.command = command
}

func (self *backupImpl) SetService(service string) {
	self.service = service
}

func (self *backupImpl) SetClient(client kubernetes.Kubernetes) {
	self.client = client
}

func (self *backupImpl) SetVolumes(volumes []corev1.PersistentVolumeClaim) {
	self.volumes = volumes
}

func (self *backupImpl) Prepare(uuid string) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	records := make([]BackupModel, 0)
	for _, volume := range self.volumes {
		records = append(records,
			BackupModel{
				BaseModel: BaseModel{
					UUID: uuid,
				},
				Namespace: self.namespace,
				Volume:    volume.Name,
				State:     PreparingBackup,
			},
		)
	}

	batchSize, err := strconv.Atoi(os.Getenv("GORM_BATCH_SIZE"))
	if err != nil {
		batchSize = 100
	}

	resp := dbConn.
		CreateInBatches(records, batchSize)
	if resp.Error == nil {
		self.uuid = uuid
	}

	return resp.Error
}

func (self *backupImpl) Backup() error {
	var rows *sql.Rows

	dbModule, err := container.Pick("elephansql")
	cnt := 0
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	if len(self.uuid) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("uuid = ?", self.uuid).
			Rows()
		if err != nil {
			return err
		}
	} else if len(self.namespace) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("namespace = ?", self.namespace).
			Rows()
	} else if len(self.volumes) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("volume in ?", getListOfVolumeName(self.volumes)).
			Rows()
	} else {
		return errors.New("Can't find approviated job")
	}

	for rows.Next() {
		var record BackupModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return err
		}

		if record.State == PreparingBackup {
			err := self.client.Do(
				fmt.Sprintf(record.UUID),
				self.command,
				self.namespace,
				self.volumes,
			)
			if err != nil {
				return err
			}

			err = self.SetStatus(record.UUID, ProvisingBackup)
			if err != nil {
				return err
			}

			cnt++
		}
	}

	if cnt > 0 {
		return nil
	} else {
		return errors.New("Can't see approviated jobs")
	}
}

func (self *backupImpl) GetStatus() (map[string]BackupState, error) {
	var rows *sql.Rows

	dbModule, err := container.Pick("elephansql")
	cnt := 0
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return nil, err
	}

	if len(self.uuid) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("uuid = ?", self.uuid).
			Rows()
		if err != nil {
			return nil, err
		}
	} else if len(self.namespace) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("namespace = ?", self.namespace).
			Rows()
	} else if len(self.volumes) > 0 {
		rows, err = dbConn.Model(&BackupModel{}).
			Where("volume in ?", getListOfVolumeName(self.volumes)).
			Rows()
	} else {
		return nil, errors.New("Can't find approviated job")
	}

	states := make(map[string]BackupState)
	for rows.Next() {
		var record BackupModel

		err = dbConn.ScanRows(rows, &record)
		if err != nil {
			return nil, err
		}

		states[record.UUID] = record.State
	}

	if cnt > 0 {
		return states, nil
	} else {
		return nil, errors.New("Can't see approviated jobs")
	}
}

func (self *backupImpl) SetStatus(uuid string, status BackupState) error {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	// @NOTE: the status must be greater than its current value.
	//        Otherwide, we must reject this update since it could
	//        cause our state machine go wary
	resp := dbConn.Model(&BackupModel{}).
		Where("uuid = ? and status < ?", uuid, status).
		Update("status", status)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func getListOfVolumeName(
	volumes []corev1.PersistentVolumeClaim,
) []string {
	ret := make([]string, 0)

	for _, volume := range volumes {
		ret = append(ret, volume.Name)
	}
	return ret
}
