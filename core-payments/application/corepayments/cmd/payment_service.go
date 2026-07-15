package cmd

import (
	"fmt"

	"github.com/moov-io/ach"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {

	achPaymentServiceCmd.PersistentFlags().StringVar(&busConnUrl, "busconnurl", "", "nats://localhost:4322")
	achPaymentServiceCmd.PersistentFlags().StringVar(&ftpServiceUrl, "ftpserviceurl", "", "localhost:6379")
	achPaymentServiceCmd.PersistentFlags().StringVar(&temporalUrl, "temporalurl", "", "localhost:7233")
	achPaymentServiceCmd.PersistentFlags().StringVar(&listenHost, "listenHost", "", "0.0.0.0")
	achPaymentServiceCmd.PersistentFlags().StringVar(&listenPort, "listenPort", "", "7080")

	fednowConnectorCmd.MarkPersistentFlagRequired("busconnurl")

	rootCmd.AddCommand(achPaymentServiceCmd)

}

var achPaymentServiceCmd = &cobra.Command{
	Use:   "payments-service",
	Short: "Start the Payments Service",
	Long: `
Start the Payments Service, which accepts payment instructions and processes them.
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Payments Service")

		fh := ach.NewFileHeader()
		fh.ImmediateDestination = "121000358" // Receiving Institutions ABA
		fh.ImmediateOrigin = "9876543219"     // Sending Company ID or ODFI
		fh.FileCreationDate = "260715"
		fh.FileCreationTime = "0729"

		bh := ach.NewBatchHeader()
		bh.ServiceClassCode = ach.MixedDebitsAndCredits
		bh.CompanyName = "Fintech Company"
		bh.CompanyIdentification = "1234567890"
		bh.StandardEntryClassCode = ach.PPD
		bh.CompanyEntryDescription = "REG.SALARY"
		bh.EffectiveEntryDate = "260716"    // need EffectiveEntryDate to be fixed so it can match output
		bh.ODFIIdentification = "121042882" // Us as ODFI or our bank

		entry := ach.NewEntryDetail()
		entry.TransactionCode = ach.CheckingCredit
		entry.SetRDFI("231380104")
		entry.DFIAccountNumber = "987654321"
		entry.Amount = 100000000
		entry.SetTraceNumber(bh.ODFIIdentification, 2)
		entry.IndividualName = "Credit Account 1"

		// build the batch
		batch := ach.NewBatchPPD(bh)
		batch.AddEntry(entry)
		batch.WithOffset(&ach.Offset{
			RoutingNumber: "121042882",
			AccountNumber: "1234567890123456",
			AccountType:   ach.OffsetChecking,
			Description:   "Payroll Checking",
		})

		if err := batch.Create(); err != nil {
			log.Fatalf("Unexpected error building batch: %s\n", err)
		}

		// build the file
		achFile := ach.NewFile()
		achFile.SetHeader(fh)
		achFile.AddBatch(batch)
		if err := achFile.Create(); err != nil {
			log.Fatalf("Unexpected error building file: %s\n", err)
		}
		fmt.Println(achFile.Header.String())
		for _, batch := range achFile.Batches {
			fmt.Println(batch.GetHeader().String())
			for _, entry := range batch.GetEntries() {
				fmt.Println(entry.String())
			}
			fmt.Println(batch.GetControl().String())
			fmt.Println(achFile.Control.String())
		}
	},
}
