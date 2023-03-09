package platform

type BackupState int

const (
	NotFoundBackup BackupState = iota
	PreparingBackup
	ProvisingBackup
	ValidatingBackup
	FinishBackup

	DefaultNamespaceToLocateBackupJob = "default"
	DefaultRsyncImageForBackup        = "alpinelinux/rsyncd"
	DefaultHeaderOfBackupHook         = "#!/bin/sh"
	DefaultMountPointOfBackupScript   = "/data/exec"
	DefaultCommandToBackupVolumes     = ""
	DefaultPreHookForBackup           = ""
	DefaultPostHookForaBackup         = ""
)

var (
	DefaultExecuterToProvisionBackup = []string{"/bin/sh", "-c", "/data/exec"}
)
