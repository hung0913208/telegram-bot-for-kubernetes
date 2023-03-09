package platform

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type Backup interface {
	Backup() error
	GetStatus() (map[string]BackupState, error)
	SetStatus(status BackupState) error
}

type backupImpl struct {
	client    kubernetes.Kubernetes
	volumes   []corev1.PersistentVolumeClaim
	hook      kubernetes.Hook
	uuid      string
	namespace string
	command   string
}

func NewBackup(
	name string,
	client kubernetes.Kubernetes,
	volumes []corev1.PersistentVolumeClaim,
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
		volumes:   volumes,
		hook:      hook,
	}
}

func NewCustomBackup(
	name, command, namespace string,
	client kubernetes.Kubernetes,
	volumes []corev1.PersistentVolumeClaim,
	hook kubernetes.Hook,
) Backup {
	return &backupImpl{
		client: kubernetes.NewFromClient(
			client,
			hook,
		),
		namespace: namespace,
		command:   command,
		volumes:   volumes,
		hook:      hook,
	}
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
				int32(0),
			)
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

func (self *backupImpl) SetStatus(status BackupState) error {
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
