package toolbox

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/clause"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/spf13/cobra"
)

type SettingToolbox interface {
	SetTimeout(timeout string)
	GetTimeout()
	SetEnable(enable string)
	GetEnable()
}

type settingToolboxImpl struct {
	toolbox *toolboxImpl
}

func newSettingToolbox(toolbox *toolboxImpl) SettingToolbox {
	return &settingToolboxImpl{
		toolbox: toolbox,
	}
}

func (self *settingToolboxImpl) SetEnable(enable string) {
	dbModule, err := container.Pick("elephansql")
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Fail get elephansql: %v", err))
		return
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Fail establish gorm: %v", err))
		return
	}

	setting := SettingModel{
		Name:  "enable",
		Type:  2,
		Value: enable,
	}

	// @NOTE: update on conflict, improve performance while keep
	//        everything safe
	dbConn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting)

	// @NOTE: save for future
	self.toolbox.enable = enable == "true"
}

func (self *settingToolboxImpl) GetEnable() {
	self.toolbox.Ok(fmt.Sprintf(
		"enable = %v",
		self.toolbox.enable,
	))
}

func (self *settingToolboxImpl) GetTimeout() {
	self.toolbox.Ok(fmt.Sprintf(
		"timeout = %d ms",
		int(self.toolbox.timeout/time.Millisecond),
	))
}

func (self *settingToolboxImpl) SetTimeout(timeout string) {
	val, err := strconv.Atoi(timeout)
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Can't accept value %s as timeout", timeout))
		return
	}

	dbModule, err := container.Pick("elephansql")
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Fail get elephansql: %v", err))
		return
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Fail establish gorm: %v", err))
		return
	}

	setting := SettingModel{
		Name:  "timeout",
		Type:  1,
		Value: timeout,
	}

	// @NOTE: update on conflict, improve performance while keep
	//        everything safe
	dbConn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting)

	// @NOTE: save for future
	self.toolbox.timeout = time.Duration(val) * time.Millisecond
}

func (self *toolboxImpl) newSettingParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "setting",
		Short: "Vercel profile",
		Long: "The Vercel profile which is used to report status " +
			"and configuration of the SRE cloud toolbox server",
	}

	root.AddCommand(&cobra.Command{
		Use:   "env",
		Short: "Print specific environment variable",
		Run: self.GenerateSafeCallback(
			"setting-env",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(fmt.Sprintf("Expect 1 argument but got %d", len(args)))
					return
				}
				self.Ok(fmt.Sprintf("%s = %s", args[0], os.Getenv(args[0])))
			},
		),
	})

	root.AddCommand(&cobra.Command{
		Use:   "timeout",
		Short: "Get/Set toolbox timeout",
		Run: self.GenerateSafeCallback(
			"setting-timeout",
			func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					newSettingToolbox(self).GetTimeout()
					return
				}

				newSettingToolbox(self).SetTimeout(args[0])
			},
		),
	})

	root.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Get/Set toolbox timeout",
		Run: self.GenerateSafeCallback(
			"setting-enable",
			func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					newSettingToolbox(self).GetEnable()
					return
				}

				newSettingToolbox(self).SetEnable(args[0])
			},
		),
	})

	return root
}
