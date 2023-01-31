package toolbox

import (
    "strconv"
    "time"
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

func (self *toolboxImpl) newVercelParser() *cobra.Command {
    root := &cobra.Command{
        Use:   "setting",
        Short: "Vercel profile",
        Long:  "The Vercel profile which is used to report status " +
               "and configuration of the SRE cloud toolbox server",
    }
    
    root.AddCommand(&cobra.Command{
        Use:   "env",
        Short: "Print specific environment variable",
        Run:   self.GenerateSafeCallback(
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
        Run:   self.GenerateSafeCallback(
            "setting-timeout",
            func(cmd *cobra.Command, args []string) {
                if len(args) == 0 {
                    self.Ok(fmt.Sprintf(
                        "timeout = %d ms",
                        int(self.timeout / time.Millisecond),
                    ))
                } else {
                    val, err := strconv.Atoi(args[0])
                    if err != nil {
                        self.Fail(fmt.Sprintf(
                            "Can't accept value %s as timeout", 
                            args[0],
                        ))
                    }

                    globalTimeout = time.Duration(val) * time.Millisecond
                }
            },
        ),
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
    root.AddCommand(self.newHealthParser())
    return root
}


