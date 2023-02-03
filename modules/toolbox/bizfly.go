package toolbox

import (
    "encoding/json"
    "fmt"

    "github.com/spf13/cobra"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
)

type BizflyToolbox interface {
    PrintAll(detail bool)
    Billing()
}

type bizflyToolboxImpl struct {
    toolbox *toolboxImpl
    client  bizfly.Api
}

func NewBizflyToolbox(toolbox *toolboxImpl, client bizfly.Api) BizflyToolbox {
    return &bizflyToolboxImpl{
        toolbox: toolbox,
        client:  client,
    }
}

func (self *bizflyToolboxImpl) Billing() {
    user, err := self.client.GetUserInfo()
    if err != nil {
        self.toolbox.Fail(fmt.Sprintf("Fail with error %v", err))
    }

    out, _ := json.Marshal(user)
    self.toolbox.Ok(fmt.Sprintf("Billing: %v", string(out)))
}

func (self *bizflyToolboxImpl) PrintAll(detail bool) {
    clusters, err := self.client.ListCluster()
    if err != nil {
        self.toolbox.Fail(fmt.Sprintf("Fail fetching clusters of %s: %v", self.client.GetAccount(), err))
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

    servers, err := self.client.ListServer()
    if err != nil {
        self.toolbox.Fail(fmt.Sprintf("Fail fetching servers of %s: %v", self.client.GetAccount(), err))
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

    volumes, err := self.client.ListVolume()
    if err != nil {
        self.toolbox.Fail(fmt.Sprintf("Fail fetching firewalls of %s: %v", self.client.GetAccount(), err))
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
        firewalls, err := self.client.ListFirewall()

        if err != nil {
            self.toolbox.Fail(fmt.Sprintf("Fail fetching firewalls of %s: %v", self.client.GetAccount(), err))
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
                        inbound.PortRange,
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

func (self *toolboxImpl) newBizflyParser() *cobra.Command {
    root := &cobra.Command{
        Use:   "bizfly",
        Short: "Bizfly cloud command line",
        Long:  "Interact with bizfly cloud throught the toolbox",
    }

    bizflyLogin := &cobra.Command{
        Use:   "login",
        Short: "Login specific bizfly account",
        Long:  "Authenticate bizfly cloud project and get token for " +
               "accessing dedicated services",
        Run:   self.GenerateSafeCallback(
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

                client, err := bizfly.NewApiWithProjectId(
                    host,
                    region,
                    project,
                    username,
                    password,
                    self.timeout,
                )
                if err != nil {
                    self.Fail(fmt.Sprintf("bizfly login fail: \n%v", err))
                    return
                }

                bizflyApi[username] = client
                if len(project) > 0 {
                    self.Ok(fmt.Sprintf("login %s, project %s, success", username, project))
                } else {
                    self.Ok(fmt.Sprintf("login %s success", username))
                }
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
        Long:  "Authenticate bizfly cloud project:\n\n" +
               "sre bizfly kubeconfig <cluster-name>",
        Run:   func(cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                return
            }
        },
    })

    root.AddCommand(&cobra.Command{
        Use:   "billing",
        Short: "List resource for each account of bizfly",
        Run:   self.GenerateSafeCallback(
            "bizfly-billing",
            func(cmd *cobra.Command, args []string) {
                for _, client := range bizflyApi {
                    bizflyToolbox := NewBizflyToolbox(self, client)
                    bizflyToolbox.Billing()
                }
            },
        ),
    })

    root.AddCommand(&cobra.Command{
        Use:   "all",
        Short: "List resource for each account of bizfly",
        Run:   self.GenerateSafeCallback(
            "bizfly-all",
            func(cmd *cobra.Command, args []string) {
                if len(bizflyApi) == 0 {
                    self.Fail("Please login bizfly before doing anything")
                }

                for username, client := range bizflyApi {
                    bizflyToolbox := NewBizflyToolbox(self, client)

                    self.Ok(fmt.Sprintf("List resource of %s", username))
                    bizflyToolbox.PrintAll(false)
                }
            },
        ),
    })

    root.AddCommand(&cobra.Command{
        Use:   "logout",
        Short: "Logout the bizfly account",
        Long:  "Logout and remove cache and resource for dedicated account:\n\n" +
               "sre bizfly logout <email>",
        Run:   self.GenerateSafeCallback(
            "bizfly-logout",
            func(cmd *cobra.Command, args []string) {
                if len(args) != 1 {
                    self.Fail(fmt.Sprintf("Expect 1 argument but got %d", len(args)))
                    return
                }

                delete(bizflyApi, args[0])
            },
        ),
    })

    root.AddCommand(bizflyLogin)
    return root
}
