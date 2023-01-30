package toolbox

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
)

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
        Run:   func(cmd *cobra.Command, args []string) {
            host, err := cmd.Flags().GetString("host")
            if err != nil {
                self.Fail(fmt.Sprintf("parse host fail: %v", err))
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
                project,
                username,
                password,
                self.timeout,
            )
            if err != nil {
                self.Fail(fmt.Sprintf("bizfly login fail: %v", err))
                return
            }

            self.bizflyApi[username] = client
            if len(project) > 0 {
                self.Ok(fmt.Sprintf("login %s, project %s, success", username, project))
            } else {
                self.Ok(fmt.Sprintf("login %s success", username))
            }
        },
    }
    bizflyLogin.PersistentFlags().
                String("host", "https://manage.bizflycloud.vn", 
                       "The bizfly control host where to manage service in dedicate regions")
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
    root.AddCommand(bizflyLogin)
    return root
}

func (self *toolboxImpl) newVercelParser() *cobra.Command {
    root := &cobra.Command{
        Use:   "vercel",
        Short: "Vercel profile",
        Long:  "The Vercel profile which is used to report status " +
               "and configuration of the SRE cloud toolbox server",
    }
    
    root.AddCommand(&cobra.Command{
        Use:   "env",
        Short: "Print specific environment variable",
        Run:   func(cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                self.Fail(fmt.Sprintf("Expect 1 argument but got %d", len(args)))
                return
            }
            self.Ok(fmt.Sprintf("%s = %s", args[0], os.Getenv(args[0])))
        },
    })
  
    return root
}

func (self *toolboxImpl) newRootParser() *cobra.Command {
    root := &cobra.Command{
        Use:   "sre",
        Short: "SRE cloud toolbox and command line",
        Long:  "The SRE command line which is used to interact with " +
               "resource of the whole company",
    }
   
    root.AddCommand(self.newBizflyParser())
    root.AddCommand(self.newVercelParser())
    return root
}


