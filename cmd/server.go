package cmd

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	cli "github.com/wayt/async/client/async"
	"github.com/wayt/async/server"
)

var (
	serverExecutorURL string
	serverBind        string
)

func init() {
	serverCmd.PersistentFlags().StringVarP(&serverExecutorURL, "executor-url", "e", fmt.Sprintf("http://127.0.0.1:%d", cli.Port), "Function executor API url")
	serverCmd.PersistentFlags().StringVarP(&serverBind, "bind", "b", ":8080", "Server bind address")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs Async server daemon",
	Run: func(cmd *cobra.Command, args []string) {
		jr := server.NewJobRepository()
		jm := server.NewJobManager(
			jr,
		)

		cr := server.NewCallbackRepository()
		cm := server.NewCallbackManager(cr)

		d := server.NewDispatcher(jm, cm, server.NewFunctionExecutor(serverExecutorURL))
		p := server.NewPoller(jr)
		go p.Poll(d)

		rescheduler := server.NewExpiredRescheduler(jm, cm)
		go rescheduler.Run()

		e := gin.Default()

		server.NewHttpJobHandler(e, jm)
		server.NewHttpCallbackHandler(e, cm, jm)

		e.Run(serverBind)
	},
}
