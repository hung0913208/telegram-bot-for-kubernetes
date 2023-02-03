package toolbox

import (
	"context"
	"fmt"
	"net/http"

	sentry "github.com/getsentry/sentry-go"
	cobra "github.com/spf13/cobra"
)

func (self *toolboxImpl) newHealthParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "health",
		Short: "Perform curl query to this API for validating performance",
		Run: self.GenerateSafeCallback("", func(cmd *cobra.Command, args []string) {
			if len(args) > 3 {
				self.Fail(fmt.Sprintf("Expect no more than 3 argument but got %d", len(args)))
				return
			}

			span := sentry.StartSpan(
				context.Background(),
				"health",
				sentry.TransactionName(args[0]),
			)

			defer func() {
				span.Finish()
			}()

			_, err := http.Get(args[1])
			if err != nil {
				self.Fail(fmt.Sprintf("Call %s fail: %v", args[1], err))
			}
		}),
	}

	return root
}
