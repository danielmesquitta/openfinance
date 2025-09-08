package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

var timeFormats = []string{
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	time.DateTime,
	time.DateOnly,
}

var (
	startDate *time.Time
	endDate   *time.Time
)

func init() {
	startOfMonth := time.Date(
		time.Now().Year(),
		time.Now().Month(),
		1,
		0,
		0,
		0,
		0,
		time.Local,
	)

	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	startDate = rootCmd.Flags().TimeP("start-date", "s", startOfMonth, timeFormats, "Start date")
	endDate = rootCmd.Flags().TimeP("end-date", "e", endOfMonth, timeFormats, "End date")
}

func Execute() {
	err := rootCmd.Execute()
	switch v := err.(type) {
	case *errs.Err:
		log.Fatalln(v, v.StackTrace)
	default:
		log.Fatalln(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "openfinance-cli",
	Short: "Open Finance Integration with Notion through Pluggy using CLI",
	Long:  "This is a tool to help you integrate your open finance data with your Notion database",
	Run:   run,
}

func run(_ *cobra.Command, _ []string) {
	syncAllUseCase := app.NewSyncAllUseCase()

	ctx := context.Background()

	err := syncAllUseCase.Execute(ctx, usecase.SyncDTO{
		StartDate: startDate.Format(time.RFC3339),
		EndDate:   endDate.Format(time.RFC3339),
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Sync completed successfully")
}
