package toolbox

import (
	"encoding/json"
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
	PrintCluster(account, project string)
	PrintPool(account, project, cluster string)
	PrintServer(account, project, clusterName string)
	PrintVolume(account, project, clusterName, server, status string)
	Sync(resource, account string)
	LinkPoolWithServer(account, project, pool string)
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
		self.toolbox.Fail("bizfly login fail: \n%v", err)
		return
	}

	if _, ok := self.toolbox.bizflyApi[username]; !ok {
		self.toolbox.bizflyApi[username] = make([]bizfly.Api, 0)
	}

	self.toolbox.bizflyApi[username] = append(
		self.toolbox.bizflyApi[username],
		client,
	)

	if len(project) > 0 {
		self.toolbox.Ok("login %s, project %s, success", username, project)
	} else {
		self.toolbox.Ok("login %s success", username)
	}
}

func (self *bizflyToolboxImpl) Billing() {
	for name, clients := range self.toolbox.bizflyApi {
		self.toolbox.Ok("Billing of %s", name)

		for _, client := range clients {
			user, err := client.GetUserInfo()
			if err != nil {
				self.toolbox.Fail("Fail with error %v", err)
			}

			out, _ := json.Marshal(user)
			self.toolbox.Ok("Billing: %v", string(out))
		}
	}
}

func (self *bizflyToolboxImpl) LinkPoolWithServer(
	account, project, pool string,
) {
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		if client.GetProjectId() != project {
			continue
		}

		pool, err := client.GetPool(pool)
		if err != nil {
			self.toolbox.Fail("Can't get list pools: %v", err)
			return
		}

		if err := client.SyncPoolNode(pool.Cluster, pool.UUID); err != nil {
			self.toolbox.Fail("Fail syncing pool %s cluster %s: %v",
				pool.UUID,
				pool.Cluster,
				err,
			)
			return
		}
	}

	self.toolbox.Ok("Sync done")
}

func (self *bizflyToolboxImpl) Sync(resource, account string) {
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		switch resource {
		case "cluster":
			clusters, err := client.ListCluster()
			if err != nil {
				self.toolbox.Fail("Can't get list clusters: %v", err)
				return
			}

			for _, clusterObj := range clusters {
				tenant, err := bizfly.NewTenant(
					client,
					clusterObj,
				)
				if err != nil {
					self.toolbox.Fail("Can't init cluster %s: %v", clusterObj.UID, err)
					ok = false
					continue
				}

				err = cluster.Join(tenant)
				if err != nil {
					self.toolbox.Fail("Can't join %s: %v", clusterObj.UID, err)
					continue
				}
			}

		case "pool":
			clusters, err := client.ListCluster()
			if err != nil {
				self.toolbox.Fail("Can't get list clusters: %v", err)
				ok = false
				return
			}

			for _, clusterObj := range clusters {
				if err := client.SyncPool(clusterObj.UID); err != nil {
					self.toolbox.Fail("Fail syncing pool of %s: %v",
						clusterObj.UID,
						err,
					)
					ok = false
					continue
				}
			}

		case "node-pool":
			clusters, err := client.ListCluster()
			if err != nil {
				self.toolbox.Fail("Can't get list clusters: %v", err)
				return
			}

			for _, clusterObj := range clusters {
				pools, err := client.ListPool(clusterObj.UID)
				if err != nil {
					self.toolbox.Fail("Can't get list pools: %v", err)
					ok = false
					continue
				}

				for _, pool := range pools {
					if err := client.SyncPoolNode(clusterObj.UID, pool.UID); err != nil {
						self.toolbox.Fail("Fail syncing pool %s cluster %s: %v",
							pool.UID,
							clusterObj.Name,
							err,
						)
						ok = false
						continue
					}
				}
			}

		case "kubernetes":
			err := client.SyncCluster()
			if err != nil {
				self.toolbox.Fail("synchronize resource `kubernetes` fail: %v", err)
				return
			}

		case "server":
			err := client.SyncServer()
			if err != nil {
				self.toolbox.Fail("synchronize resource `server` fail: %v", err)
				return
			}

		case "volume":
			err := client.SyncVolume()
			if err != nil {
				self.toolbox.Fail("synchronize resource `volume` fail: %v", err)
				return
			}

		case "firewall":
			err := client.SyncFirewall()
			if err != nil {
				self.toolbox.Fail("synchronize resource `firewall` fail: %v", err)
				return
			}

		default:
			self.toolbox.Fail("Don't support resource `%s`", resource)
			return
		}
	}

	if ok {
		self.toolbox.Ok("Sync done")
	}
}

func (self *bizflyToolboxImpl) PrintCluster(account, project string) {
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		if client.GetProjectId() != project {
			continue
		}

		clusters, err := client.ListCluster()
		if err != nil {
			self.toolbox.Fail("Can't get list clusters: %v", err)
			return
		}

		for _, cluster := range clusters {
			self.toolbox.Ok(
				" +  %s - %s - %s - %v",
				cluster.UID,
				cluster.Name,
				cluster.ClusterStatus,
				cluster.Tags,
			)
		}
		return
	}
}

func (self *bizflyToolboxImpl) PrintPool(account, project, clusterName string) {
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		if client.GetProjectId() != project {
			continue
		}

		clusters, err := client.ListCluster()
		if err != nil {
			self.toolbox.Fail("Can't get list clusters: %v", err)
			return
		}

		for _, clusterObj := range clusters {
			if len(clusterName) > 0 && clusterObj.Name != clusterName {
				continue
			}

			pools, err := client.ListPool(clusterObj.UID)
			if err != nil {
				self.toolbox.Fail("Can't get list pools: %v", err)
				return
			}

			for _, pool := range pools {
				self.toolbox.Ok(
					" +  %s - %s - %s",
					pool.UID,
					clusterObj.Name,
					pool.ProvisionStatus,
				)
			}
		}
	}
}

func (self *bizflyToolboxImpl) PrintServer(account, project, clusterName string) {
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		if client.GetProjectId() != project {
			continue
		}

		clusters, err := client.ListCluster()
		if err != nil {
			self.toolbox.Fail("Can't get list clusters: %v", err)
			return
		}

		if len(clusterName) == 0 {
			servers, err := client.ListServer()
			if err != nil {
				self.toolbox.Fail("Can't get list pools: %v", err)
				return
			}

			for _, server := range servers {
				self.toolbox.Ok(
					" +  %s - %s",
					server.ID,
					server.Status,
				)
			}
		} else {
			for _, clusterObj := range clusters {
				if len(clusterName) > 0 && clusterObj.Name != clusterName {
					continue
				}

				servers, err := client.ListServer(clusterObj.UID)
				if err != nil {
					self.toolbox.Fail("Can't get list pools: %v", err)
					return
				}

				for _, server := range servers {
					self.toolbox.Ok(
						" +  %s - %s",
						server.ID,
						server.Status,
					)
				}
			}
		}
	}
}

func (self *bizflyToolboxImpl) PrintVolume(
	account, project, clusterName, serverName, status string,
) {
	cnt := 0
	clients, ok := self.toolbox.bizflyApi[account]
	if !ok {
		self.toolbox.Fail("Unknown %s", account)
		return
	}

	for _, client := range clients {
		if client.GetProjectId() != project {
			continue
		}

		clusters, err := client.ListCluster()
		if err != nil {
			self.toolbox.Fail("Can't get list clusters: %v", err)
			return
		}

		if len(clusterName) == 0 {
			servers, err := client.ListServer()
			if err != nil {
				self.toolbox.Fail("Can't get list pools: %v", err)
				return
			}

			for _, server := range servers {
				if len(serverName) > 0 && server.Name != serverName {
					continue
				}

				volumes, err := client.ListVolume(server.ID)
				if err != nil {
					self.toolbox.Fail("Can't get list volumes: %v", err)
					return
				}

				for _, volume := range volumes {
					if len(status) > 0 && status != volume.Status {
						continue
					}

					self.toolbox.Ok(
						" +  %s - %s",
						volume.ID,
						volume.Status,
					)
					cnt += 1
					if cnt > 10 {
						self.toolbox.Flush()
						cnt = 0
					}
				}
			}
		} else {
			for _, clusterObj := range clusters {
				if len(clusterName) > 0 && clusterObj.Name != clusterName {
					continue
				}

				servers, err := client.ListServer(clusterObj.UID)
				if err != nil {
					self.toolbox.Fail("Can't get list pools: %v", err)
					return
				}

				for _, server := range servers {
					if len(serverName) > 0 && server.Name != serverName {
						continue
					}

					volumes, err := client.ListVolume(server.ID)
					if err != nil {
						self.toolbox.Fail("Can't get list volumes: %v", err)
						return
					}
					for _, volume := range volumes {
						if len(status) > 0 && status != volume.Status {
							continue
						}

						self.toolbox.Ok(
							" +  %s - %s",
							volume.ID,
							volume.Status,
						)
						cnt += 1
						if cnt > 10 {
							self.toolbox.Flush()
							cnt = 0
						}
					}
				}
			}
		}
	}
}

func (self *bizflyToolboxImpl) PrintAll(detail bool) {
	for name, clients := range self.toolbox.bizflyApi {
		self.toolbox.Ok("Detail info of %s", name)

		for _, client := range clients {
			volumes, err := client.ListVolume()
			if err != nil {
				self.toolbox.Fail("Fail fetching firewalls of %s: %v", client.GetAccount(), err)
			}

			servers, err := client.ListServer()
			if err != nil {
				self.toolbox.Fail("Fail fetching servers of %s: %v", client.GetAccount(), err)
			}

			clusters, err := client.ListCluster()
			if err != nil {
				self.toolbox.Fail("Fail fetching clusters of %s: %v", client.GetAccount(), err)
			}

			if len(clusters) > 0 {
				self.toolbox.Ok("List clusters:\n")

				for _, cluster := range clusters {
					self.toolbox.Ok(
						" + %s - %s - %s - %v",
						cluster.UID,
						cluster.Name,
						cluster.ClusterStatus,
						cluster.Tags,
					)
				}

				self.toolbox.Flush()
			}

			if len(servers) > 0 {
				self.toolbox.Ok("List servers:\n")

				for _, server := range servers {
					self.toolbox.Ok(
						" + %s - %s - %s",
						server.ID,
						server.Name,
						server.Status,
					)

					if detail {
						for _, vol := range server.AttachedVolumes {
							self.toolbox.Ok(
								"   \\-> %s - %s",
								vol.ID,
								vol.Type,
							)
						}
					}
				}
				self.toolbox.Flush()
			}

			if len(volumes) > 0 {
				self.toolbox.Ok("List volumes:\n")

				for _, vol := range volumes {
					self.toolbox.Ok(
						" + %s - %v - %v",
						vol.ID,
						vol.Status,
						vol.VolumeType,
					)
				}
				self.toolbox.Flush()
			}

			if detail {
				if len(servers) > 0 {
					self.toolbox.Ok("List servers:\n")

					for _, server := range servers {
						self.toolbox.Ok(
							" + %s - %s - %s",
							server.ID,
							server.Name,
							server.Status,
						)

						if detail {
							for _, vol := range server.AttachedVolumes {
								self.toolbox.Ok(
									"   \\-> %s - %s",
									vol.ID,
									vol.Type,
								)
							}
						}
					}
					self.toolbox.Flush()
				}

				if len(volumes) > 0 {
					self.toolbox.Ok("List volumes:\n")

					for _, vol := range volumes {
						self.toolbox.Ok(
							" + %s - %v - %v",
							vol.ID,
							vol.Status,
							vol.VolumeType,
						)
					}
					self.toolbox.Flush()
				}

				firewalls, err := client.ListFirewall()

				if err != nil {
					self.toolbox.Fail(
						"Fail fetching firewalls of %s: %v",
						client.GetAccount(),
						err,
					)
				}

				if len(firewalls) > 0 {
					self.toolbox.Ok("List firewalls:\n")

					for _, firewall := range firewalls {
						self.toolbox.Ok(
							" + %s - %v",
							firewall.ID,
							firewall.Tags,
						)

						for _, inbound := range firewall.InBound {
							self.toolbox.Ok(
								"   >>> %s - %s : { %s }",
								inbound.ID,
								inbound.Tags,
								inbound.CIDR,
							)
						}

						for _, outbound := range firewall.InBound {
							self.toolbox.Ok(
								"   <<< %s - %s : { %s }",
								outbound.ID,
								outbound.CIDR,
								outbound.PortRange,
							)
						}
					}
					self.toolbox.Flush()
				}
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

	bizflySyncGroupCmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize resource between cloud and toolbox",
	}

	bizflyPrintGroupCmd := &cobra.Command{
		Use:   "print",
		Short: "Print resource for dedicated IAM",
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
					self.Fail("parse host fail: %v", err)
					return
				}
				region, err := cmd.Flags().GetString("region")
				if err != nil {
					self.Fail("parse region fail: %v", err)
					return
				}
				project, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}
				username, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}
				password, err := cmd.Flags().GetString("password")
				if err != nil {
					self.Fail("parse password fail: %v", err)
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
	bizflyLinkPoolWithServerCmd := &cobra.Command{
		Use:   "node-pool",
		Short: "Synchronize resource `node-pool` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-node-pool",
			func(cmd *cobra.Command, args []string) {
				projectId, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}

				account, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}

				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).LinkPoolWithServer(account, projectId, args[0])
			},
		),
	}

	bizflyPrintClusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Print all cluster of dedicated IAM",
		Run: self.GenerateSafeCallback(
			"bizfly-print-cluster",
			func(cmd *cobra.Command, args []string) {
				projectId, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}

				account, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}

				newBizflyToolbox(self).PrintCluster(account, projectId)
			},
		),
	}

	bizflyPrintPoolCmd := &cobra.Command{
		Use:   "pool",
		Short: "Print all pool of specific cluster of dedicated IAM",
		Run: self.GenerateSafeCallback(
			"bizfly-print-pool",
			func(cmd *cobra.Command, args []string) {
				projectId, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}

				account, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}

				cluster, err := cmd.Flags().GetString("cluster")
				if err != nil {
					self.Fail("parse cluster fail: %v", err)
					return
				}

				newBizflyToolbox(self).PrintPool(account, projectId, cluster)
			},
		),
	}

	bizflyPrintServerCmd := &cobra.Command{
		Use:   "server",
		Short: "Print all server of dedicated IAM",
		Run: self.GenerateSafeCallback(
			"bizfly-print-server",
			func(cmd *cobra.Command, args []string) {
				projectId, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}

				account, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}

				cluster, err := cmd.Flags().GetString("cluster")
				if err != nil {
					self.Fail("parse cluster fail: %v", err)
					return
				}

				newBizflyToolbox(self).PrintServer(account, projectId, cluster)
			},
		),
	}

	bizflyPrintVolumeCmd := &cobra.Command{
		Use:   "volume",
		Short: "Print all volume of dedicated IAM",
		Run: self.GenerateSafeCallback(
			"bizfly-print-volume",
			func(cmd *cobra.Command, args []string) {
				projectId, err := cmd.Flags().GetString("project-id")
				if err != nil {
					self.Fail("parse project-id fail: %v", err)
					return
				}

				account, err := cmd.Flags().GetString("email")
				if err != nil {
					self.Fail("parse email fail: %v", err)
					return
				}

				cluster, err := cmd.Flags().GetString("cluster")
				if err != nil {
					self.Fail("parse cluster fail: %v", err)
					return
				}

				server, err := cmd.Flags().GetString("server")
				if err != nil {
					self.Fail("parse server fail: %v", err)
					return
				}
				status, err := cmd.Flags().GetString("status")
				if err != nil {
					self.Fail("parse status fail: %v", err)
					return
				}

				newBizflyToolbox(self).
					PrintVolume(account, projectId, cluster, server, status)
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

	bizflyLinkPoolWithServerCmd.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyLinkPoolWithServerCmd.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")

	bizflyPrintClusterCmd.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyPrintClusterCmd.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")

	bizflyPrintPoolCmd.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyPrintPoolCmd.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")
	bizflyPrintPoolCmd.PersistentFlags().
		String("cluster", "",
			"The cluster id which identify which is used to filter output")

	bizflyPrintServerCmd.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyPrintServerCmd.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")
	bizflyPrintServerCmd.PersistentFlags().
		String("cluster", "",
			"The server id which identify which is used to filter output")

	bizflyPrintVolumeCmd.PersistentFlags().
		String("email", "", "The email which is used to identify accounts")
	bizflyPrintVolumeCmd.PersistentFlags().
		String("project-id", "",
			"The project id which is used to identify and isolate billing resource")
	bizflyPrintVolumeCmd.PersistentFlags().
		String("cluster", "",
			"The cluster id which identify which is used to filter output")
	bizflyPrintVolumeCmd.PersistentFlags().
		String("server", "",
			"The server id which identify which is used to filter output")
	bizflyPrintVolumeCmd.PersistentFlags().
		String("status", "",
			"The expected status of volumes")

	root.AddCommand(&cobra.Command{
		Use:   "billing",
		Short: "Show billing report",
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
				newBizflyToolbox(self).PrintAll(len(args) > 0 && args[0] == "true")
			},
		),
	})

	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "cluster",
		Short: "Synchronize resource `cluster` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-cluster",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("cluster", args[0])
			},
		),
	})
	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Synchronize resource `server` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-server",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("server", args[0])
			},
		),
	})
	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "pool",
		Short: "Synchronize resource `pool` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-pool",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("pool", args[0])
			},
		),
	})
	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "kubernetes",
		Short: "Synchronize resource `cluster` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-kubernetes",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("kubernetes", args[0])
			},
		),
	})
	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "volume",
		Short: "Synchronize resource `cluster` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-volume",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("volume", args[0])
			},
		),
	})
	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "firewall",
		Short: "Synchronize resource `cluster` between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-firewall",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("firewall", args[0])
			},
		),
	})

	bizflySyncGroupCmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Synchronize all resource between cloud and toolbox",
		Run: self.GenerateSafeCallback(
			"bizfly-sync-all",
			func(cmd *cobra.Command, args []string) {
				if len(args) != 1 {
					self.Fail(
						"Sync requires 1 arguments, you have %d",
						len(args),
					)
					return
				}

				newBizflyToolbox(self).Sync("kubernetes", args[0])
				newBizflyToolbox(self).Sync("server", args[0])
				newBizflyToolbox(self).Sync("volume", args[0])
				newBizflyToolbox(self).Sync("firewall", args[0])
				newBizflyToolbox(self).Sync("cluster", args[0])
				newBizflyToolbox(self).Sync("firewall", args[0])
				newBizflyToolbox(self).Sync("pool", args[0])
				newBizflyToolbox(self).Sync("node-pool", args[0])
			},
		),
	})

	bizflyPrintGroupCmd.AddCommand(bizflyPrintClusterCmd)
	bizflyPrintGroupCmd.AddCommand(bizflyPrintPoolCmd)
	bizflyPrintGroupCmd.AddCommand(bizflyPrintServerCmd)
	bizflyPrintGroupCmd.AddCommand(bizflyPrintVolumeCmd)

	bizflySyncGroupCmd.AddCommand(bizflyLinkPoolWithServerCmd)
	root.AddCommand(bizflySyncGroupCmd)
	root.AddCommand(bizflyPrintGroupCmd)
	root.AddCommand(bizflyLogin)
	return root
}
