package test

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
)

func TestKubernetesSimple(t *testing.T) {
	dbcli, _ := db.NewPg(
		"tiny.db.elephantsql.com",
		5432,
		"dqihagqk",
		"ur5VTr-fIvj1SF5m491_TmtX_zWtO7y3",
		"dqihagqk",
		time.Duration(1000)*time.Millisecond)
	dbConn := dbcli.Establish()

	rows, _ := dbConn.Model(&cluster.ClusterModel{}).Rows()
	defer rows.Close()

	for rows.Next() {
		var record cluster.ClusterModel

		_ = dbConn.ScanRows(rows, &record)

		if len(record.Name) == 0 {
			break
		}

		kubeconfig, err := base64.StdEncoding.DecodeString(record.Kubeconfig)
		if err != nil {
			continue
		}

		dat, err := os.ReadFile("/tmp/kubeconfig.txt")
		fmt.Printf("%x\n", md5.Sum(dat))
		fmt.Printf("%x\n", md5.Sum(kubeconfig))

		tenant, err := kubernetes.NewDefaultTenant(
			record.Name,
			kubeconfig,
		)

		fmt.Println(err)

		cli, _ := tenant.GetClient()
		cli.GetPods()
	}
}
