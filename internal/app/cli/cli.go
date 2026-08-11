package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase/ingest"
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

const (
	monthFlag     = "month"
	yearFlag      = "year"
	startDateFlag = "start-date"
	endDateFlag   = "end-date"
)

func init() {
	now := time.Now()

	rootCmd.Flags().IntP(monthFlag, "m", int(now.Month()), "Month (1-12)")
	rootCmd.Flags().IntP(yearFlag, "y", now.Year(), "Year (YYYY)")

	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	rootCmd.Flags().TimeP(startDateFlag, "s", startOfMonth, timeFormats, "Start date")
	rootCmd.Flags().TimeP(endDateFlag, "e", endOfMonth, timeFormats, "End date")
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "openfinance-cli",
	Short: "Open Finance Integration with Notion through Pluggy using CLI",
	Long:  "This is a tool to help you integrate your open finance data with your Notion database",
	RunE:  run,
}

func run(cmd *cobra.Command, _ []string) error {
	ingestUseCase, err := app.NewIngestUseCase()
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}

	return executeIngest(cmd, ingestUseCase)
}

func executeIngest(cmd *cobra.Command, ingestUseCase ingest.IngestExecutor) error {
	monthVal, _ := cmd.Flags().GetInt(monthFlag)
	yearVal, _ := cmd.Flags().GetInt(yearFlag)
	startDateVal, _ := cmd.Flags().GetTime(startDateFlag)
	endDateVal, _ := cmd.Flags().GetTime(endDateFlag)

	if cmd.Flags().Changed(monthFlag) || cmd.Flags().Changed(yearFlag) {
		month := time.Month(monthVal)
		year := yearVal

		startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		startDateVal = startOfMonth
		endDateVal = endOfMonth
	}

	ctx := context.Background()

	err := ingestUseCase.Execute(ctx, ingest.IngestInput{
		StartDate: startDateVal,
		EndDate:   endDateVal,
	})
	if err != nil {
		return fmt.Errorf("execute ingest: %w", err)
	}

	if _, err := fmt.Println("Ingest completed successfully"); err != nil {
		return fmt.Errorf("print success message: %w", err)
	}

	return nil
}
