package core_payments_test

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/google/uuid"
	"github.com/moov-io/fednow20022/gen/pacs_008_001_08"
	"github.com/moov-io/fednow20022/pkg/fednow"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FednowInbound", func() {

	var js jetstream.JetStream

	BeforeEach(func() {

		nc, _ := nats.Connect("localhost:4222")
		js, _ = jetstream.New(nc)

	})

	Describe("Receiving and processing an inbound customer credit transfer Fednow message", func() {

		message := pacs_008_001_08.Document{}

		It("Creates a fednow customer credit transfer message and submits to queue", func() {

			msgId := fednow.MessageID(time.Now(), "123456789", "98765432109876543")
			controlSum := pacs_008_001_08.DecimalNumber(float64(100.00))
			totalSettlementAmount := pacs_008_001_08.ActiveCurrencyAndAmount{
				Ccy:  "USD",
				Text: "100.00",
			}
			settlementDate := fednow.ISODate(time.Now().AddDate(0, 0, -1))
			clearingSystem := pacs_008_001_08.ExternalCashClearingSystem1Code("FDW")

			ultimateDebtorName := pacs_008_001_08.Max140Text("Ultimate Debtorperson")
			ultimateCreditorName := pacs_008_001_08.Max140Text("Ultimate Creditorperson")

			message.FIToFICstmrCdtTrf = pacs_008_001_08.FIToFICustomerCreditTransferV08{
				GrpHdr: pacs_008_001_08.GroupHeader93{
					MsgId:             pacs_008_001_08.Max35Text(msgId),
					CreDtTm:           fednow.ISODateTime(time.Now()),
					BtchBookg:         nil,
					NbOfTxs:           "1",
					CtrlSum:           &controlSum,
					TtlIntrBkSttlmAmt: &totalSettlementAmount,
					IntrBkSttlmDt:     &settlementDate,
					SttlmInf: pacs_008_001_08.SettlementInstruction7{
						SttlmMtd: "CLRG",
						ClrSys: &pacs_008_001_08.ClearingSystemIdentification3Choice{
							Cd: &clearingSystem,
						},
					},
				},
				CdtTrfTxInf: []pacs_008_001_08.CreditTransferTransaction39{
					{
						PmtId: pacs_008_001_08.PaymentIdentification7{
							EndToEndId: pacs_008_001_08.Max35Text(uuid.NewString()),
						},
						PmtTpInf: nil,
						IntrBkSttlmAmt: pacs_008_001_08.ActiveCurrencyAndAmount{
							Ccy:  "USD",
							Text: "100.00",
						},
						IntrBkSttlmDt: &settlementDate,
						UltmtDbtr: &pacs_008_001_08.PartyIdentification135{
							Nm: &ultimateDebtorName,
							PstlAdr: &pacs_008_001_08.PostalAddress24{
								XMLName:     xml.Name{},
								AdrTp:       nil,
								Dept:        nil,
								SubDept:     nil,
								StrtNm:      nil,
								BldgNb:      nil,
								BldgNm:      nil,
								Flr:         nil,
								PstBx:       nil,
								Room:        nil,
								PstCd:       nil,
								TwnNm:       nil,
								TwnLctnNm:   nil,
								DstrctNm:    nil,
								CtrySubDvsn: nil,
								Ctry:        nil,
								AdrLine:     nil,
							},
							Id:        nil,
							CtryOfRes: nil,
							CtctDtls:  nil,
						},
						InitgPty:    nil,
						Dbtr:        pacs_008_001_08.PartyIdentification135{},
						DbtrAcct:    nil,
						DbtrAgt:     pacs_008_001_08.BranchAndFinancialInstitutionIdentification6{},
						DbtrAgtAcct: nil,
						CdtrAgt:     pacs_008_001_08.BranchAndFinancialInstitutionIdentification6{},
						CdtrAgtAcct: nil,
						Cdtr:        pacs_008_001_08.PartyIdentification135{},
						CdtrAcct:    nil,
						UltmtCdtr: &pacs_008_001_08.PartyIdentification135{
							Nm:        &ultimateCreditorName,
							PstlAdr:   nil,
							Id:        nil,
							CtryOfRes: nil,
							CtctDtls:  nil,
						},
						Purp:        nil,
						RgltryRptg:  nil,
						Tax:         nil,
						RltdRmtInf:  nil,
						RmtInf:      nil,
						SplmtryData: nil,
					},
				},
				SplmtryData: nil,
			}

			// Marshall it
			rawMessage, err := xml.Marshal(&message)
			Expect(err).To(BeNil())
			Expect(rawMessage).To(Not(BeNil()))

			// Submit to queue!
			ack, err := js.Publish(context.Background(), "FEDNOWIN.inbound", rawMessage)
			Expect(err).To(BeNil())
			Expect(ack).ToNot(BeNil())

		})

	})

})
