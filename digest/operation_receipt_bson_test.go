package digest

import (
	"testing"
	"time"

	"github.com/imfact-labs/currency-model/common"
	cdigest "github.com/imfact-labs/currency-model/digest"
	mongodbst "github.com/imfact-labs/currency-model/digest/mongodb"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	jsonenc "github.com/imfact-labs/mitum2/util/encoder/json"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestContractOperationDocsPreserveCurrencyReceipt(t *testing.T) {
	encs, benc := newReceiptDigestTestEncoders(t)
	sender := base.NewStringAddress("digestreceiptsender")
	target := base.NewStringAddress("digestreceipttarget")
	currency := ctypes.CurrencyID("FEE")

	register, err := contract.NewRegisterContract(contract.NewRegisterContractFact(
		[]byte("register"), sender, target, "package contract", map[string]string{"value": "one"}, currency,
	))
	if err != nil {
		t.Fatalf("new register operation: %v", err)
	}
	call, err := contract.NewCallContract(contract.NewCallContractFactWithItems(
		[]byte("call"), sender, target,
		[]contract.CallContractItem{contract.NewCallContractItem("Set", map[string]string{"value": "two"})},
		currency,
	))
	if err != nil {
		t.Fatalf("new call operation: %v", err)
	}
	privatekey := base.NewMPrivatekey()
	networkID := base.NetworkID("digest-receipt-test")
	if err := register.Sign(privatekey, networkID); err != nil {
		t.Fatalf("sign register operation: %v", err)
	}
	if err := call.Sign(privatekey, networkID); err != nil {
		t.Fatalf("sign call operation: %v", err)
	}

	for _, op := range []base.Operation{register, call} {
		t.Run(op.Hint().Type().String(), func(t *testing.T) {
			receipt := ctypes.NewCurrencyOperationReceipt(
				ctypes.FixedFeeerHint.String(),
				ctypes.NewFixedFeeReceipt(currency, common.NewBig(7)),
				nil,
			)
			doc, err := cdigest.NewOperationDoc(
				op, benc, base.Height(3), time.Unix(123, 0).UTC(), true, "", 0, receipt,
			)
			if err != nil {
				t.Fatalf("new operation doc: %v", err)
			}

			body, err := doc.MarshalBSON()
			if err != nil {
				t.Fatalf("marshal operation doc: %v", err)
			}
			assertReceiptHintInOperationDoc(t, body)

			_, value, err := mongodbst.LoadDataFromDoc(body, encs)
			if err != nil {
				t.Fatalf("decode operation doc: %v", err)
			}
			got, ok := value.(cdigest.OperationValue)
			if !ok {
				t.Fatalf("decoded operation value type = %T", value)
			}
			if !got.Operation().Hash().Equal(op.Hash()) || !got.Operation().Fact().Hash().Equal(op.Fact().Hash()) {
				t.Fatal("operation or fact hash changed during digest BSON round trip")
			}
			gotReceipt, ok := got.Receipt().(ctypes.CurrencyOperationReceipt)
			if !ok {
				t.Fatalf("decoded receipt type = %T", got.Receipt())
			}
			fee, ok := gotReceipt.Fee.(ctypes.FixedFeeReceipt)
			if !ok {
				t.Fatalf("decoded fee receipt type = %T", gotReceipt.Fee)
			}
			if fee.Currency() != currency || fee.FeeAmount() != "7" {
				t.Fatalf("decoded fee = %q %q", fee.Currency(), fee.FeeAmount())
			}
		})
	}
}

func newReceiptDigestTestEncoders(t *testing.T) (*encoder.Encoders, *bsonenc.Encoder) {
	t.Helper()
	jenc := jsonenc.NewEncoder()
	encs := encoder.NewEncoders(jenc, jenc)
	benc := bsonenc.NewEncoder()
	if err := encs.AddEncoder(benc); err != nil {
		t.Fatalf("add BSON encoder: %v", err)
	}
	if err := steps.LoadHinters(encs); err != nil {
		t.Fatalf("load hinters: %v", err)
	}
	return encs, benc
}

func assertReceiptHintInOperationDoc(t *testing.T, body []byte) {
	t.Helper()
	var outer struct {
		Data bson.Raw `bson:"d"`
	}
	if err := bson.Unmarshal(body, &outer); err != nil {
		t.Fatalf("decode outer operation doc: %v", err)
	}
	var value struct {
		Receipt bson.Raw `bson:"receipt"`
	}
	if err := bson.Unmarshal(outer.Data, &value); err != nil {
		t.Fatalf("decode operation value: %v", err)
	}
	if got := value.Receipt.Lookup("_hint"); got.Type != bson.TypeString || got.StringValue() != ctypes.CurrencyOperationReceiptHint.String() {
		t.Fatalf("receipt _hint = (%s, %q)", got.Type, got.StringValue())
	}
}
