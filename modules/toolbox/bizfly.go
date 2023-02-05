package toolbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/cluster"
	"github.com/spf13/cobra"
)

type BizflyToolbox interface {
	Login(
		host, region, project, username, password string,
		timeout time.Duration,
	)
	PrintAll(detail bool)
	Sync(account string)
	Billing()
}

type bizflyToolboxImpl struct {
	toolbox *toolboxImpl
}

func newBizflyToolbox(toolbox *toolboxImpl) BizflyToolbox {
	return &bizflyToolboxImpl{
		toolbox: toolbox,
	}
}

func (self *bizflyToolboxImpl) Login(
	host, region, project, username, password string,
	timeout time.Duration,
) {
	client, err := bizfly.NewApiWithProjectId(
		host,
		region,
		project,
		username,
		password,
		timeout,
	)
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("bizfly login fail: \n%v", err))
		return
	}

	self.toolbox.bizflyApi[username] = client

	if len(project) > 0 {
		self.toolbox.Ok("login %s, project %s, success", username, project)
	} else {
		self.toolbox.Ok("login %s success", username)
	}
}

func (self *bizflyToolboxImpl) Billing() {
	for name, client := range self.toolbox.bizflyApi {
		user, err := client.GetUserInfo()
		if err != nil {
			self.toolbox.Fail(fmt.Sprintf("Fail with error %v", err))
		}

		self.toolbox.Ok(fmt.Sprintf("Billing of %s", name))

		out, _ := json.Marshal(user)
		self.toolbox.Ok(fmt.Sprintf("Billing: %v", string(out)))
	}
}

func (self *bizflyToolboxImpl) Sync(account string) {
	client, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail(fmt.Sprintf("Unknown %s", account))
		return
	}

	clusters, err := client.ListCluster()
	if err != nil {
		self.toolbox.Fail(fmt.Sprintf("Can't get list clusters: %v", err))
		return
	}

	for _, clusterObj := range clusters {
		tenant, err := bizfly.NewTenant(
			client,
			clusterObj,
		)
		if err != nil {
			self.toolbox.Fail("Can't init cluster %s: %v", clusterObj.UID, err)
			continue
		}

		err = cluster.Join(tenant)
		if err != nil {
			self.toolbox.Fail("Can't join %s: %v", clusterObj.UID, err)
		}
	}
}

func (self *bizflyToolboxImpl) PrintAll(detail bool) {
	for name, client := range self.toolbox.bizflyApi {
		self.toolbox.Ok(fmt.Sprintf("Detail info of %s", name))

		volumes, err := client.ListVolume()
		if err != nil {
			self.toolbox.Fail(fmt.Sprintf("Fail fetching firewalls of %s: %v", client.GetAccount(), err))
		}

		servers, err := client.ListServer()
		if err != nil {
			self.toolbox.Fail(fmt.Sprintf("Fail fetching servers of %s: %v", client.GetAccount(), err))
		}

		clusters, err := client.ListCluster()
		if err != nil {
			self.toolbox.Fail(fmt.Sprintf("Fail fetching clusters of %s: %v", client.GetAccount(), err))
		}

		if len(clusters) > 0 {
			self.toolbox.Ok("List clusters:\n")

			for _, cluster := range clusters {
				self.toolbox.Ok(fmt.Sprintf(
					" + %s - %s - %s - %v",
					cluster.UID,
					cluster.Name,
					cluster.ClusterStatus,
					cluster.Tags,
				))
			}

			self.toolbox.Flush()
		}

		if len(servers) > 0 {
			self.toolbox.Ok("List servers:\n")

			for _, server := range servers {
				self.toolbox.Ok(fmt.Sprintf(
					" + %s - %s - %s",
					server.ID,
					server.Name,
					server.Status,
				))

				if detail {
					for _, vol := range server.AttachedVolumes {
						self.toolbox.Ok(fmt.Sprintf(
							"   \\-> %s - %s",
							vol.ID,
							vol.Type,
						))
					}
				}
			}
			self.toolbox.Flush()
		}

		if len(volumes) > 0 {
			self.toolbox.Ok("List volumes:\n")

			for _, vol := range volumes {
				self.toolbox.Ok(fmt.Sprintf(
					" + %s - %v - %v",
					vol.ID,
					vol.Status,
					vol.VolumeType,
				))
			}
			self.toolbox.Flush()
		}

		if detail {
			firewalls, err := client.ListFirewall()

			if err != nil {
				self.toolbox.Fail(fmt.Sprintf("Fail fetching firewalls of %s: %v", client.GetAccount(), err))
			}

			if len(firewalls) > 0 {
				self.toolbox.Ok("List firewalls:\n")

				for _, firewall := range firewalls {
					self.toolbox.Ok(fmt.Sprintf(
						" + %s - %v",
						firewall.ID,
						firewall.Tags,
					))

					for _, inbound := range firewall.InBound {
						self.toolbox.Ok(fmt.Sprintf(
							"   >>> %s - %s : { %s }",
							inbound.ID,
							inbound.Tags,
							inbound.CIDR,
						))
					}

					for _, outbound := range firewall.InBound {
						self.toolbox.Ok(fmt.Sprintf(
							"   <<< %s - %s : { %s }",
							outbound.ID,
							outbound.CIDR,
							outbound.PortRange,
						))
					}
				}
				self.toolbox.Flush()
			}
		}
	}
}

func (self *toolboxImpl) newBizflyParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "bizfly",
		Short: "Bizfly cloud command line",
		Long:  "Interact with bizfly cloud throught the toolbox",
	}

	bizflyLogin := &cobra.Command{
		Use:   "login",
		Short: "Login specific bizfly account",
		Long: "Authenticate bizfly cloud project and get token for " +
			"accessing dedicated services",
		Run: self.GenerateSafeCallback(
			"bizfly-login",
			func(cmd *cobra.Command, args []string) {
				host, err := cmd.Flags().GetString("host")
				if err != nil {
					self.Fail(fmt.Sprintf("parse host fail: %v", err))
					return
				}
				region, err := cmd.Flags().GetString("region")
				if err != nil {
					self.Fail(fmt.Sprintf("parse region fail: %v", err))
					return
				}
				project, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail(fmt.Sprintf("parse project-id fail: %v", err))
					return
				}
				username, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail(fmt.Sprintf("parse email fail: %v", err))
					return
				}
				password, err := cmd.Flags().GetString("password")
				if err != nil {
					self.Fail(fmt.Sprintf("parse password fail: %v", err))
					return
				}

				newBizflyToolbox(self).
					Login(
						host,
						region,
						project,
						username,
						password,
						self.timeout,
					)
			},
		),
	}
	bizflyLogin.PersistentFlags().
		String("host", "https://manage.bizflycloud.vn",
			"The bizfly control host where to manage service in dedicate regions")
	bizflyLogin.PersistentFlags().String("region", "HN", "The bizfly region")
	bizflyLogin.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyLogin.PersistentFlags().
		String("password", "", "The password of this account")
	bizflyLogin.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")

	root.AddCommand(&cobra.Command{
		Use:   "kubeconfig",
		Short: "Login specific bizfly account",
		Long: "Authenticate bizfly cloud project:\n\n" +
			"sre bizfly kubeconfig <cluster-name>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				return
			}
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "billing",
		Short: "List resource for each account of bizfly",
		Run: self.GenerateSafeCallback(
			"bizfly-billing",
			func(cmd *cobra.Command, args []string) {
				newBizflyToolbox(self).Billing()
			},
		),
	})

	root.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "List resource for each account of bizfly",
		Run: self.GenerateSafeCallback(
			"bizfly-all",
			func(cmd *cobra.Command, args []string) {
				newBizflyToolbox(self).PrintAll(false)
			},
		),
	})

	root.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "List resource for each account of bizfly",
		Run: self.GenerateSafeCallback(
			"bizfly-sync",
			func(cmd *cobra.Command, args []string) {
				newBizflyToolbox(self).Sync(args[0])
			},
		),
	})
	root.AddCommand(bizflyLogin)
	return root
}
