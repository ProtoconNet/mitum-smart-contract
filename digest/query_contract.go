package digest

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"net/http"

	cdigest "github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
	pkgerrors "github.com/pkg/errors"
)

var digestContractQueryEngine runtime.ContractEngine = runtime.NewGnoEngine()

const MaxContractQueryBodyBytes = 128 * 1024

type ContractQueryResponse struct {
	Contract string              `json:"contract"`
	Function string              `json:"function"`
	Engine   string              `json:"engine"`
	ReadOnly bool                `json:"read_only"`
	Output   ContractQueryOutput `json:"output"`
}

type ContractQueryOutput struct {
	Result interface{} `json:"result"`
	Ok     *bool       `json:"ok,omitempty"`
}

func (hd *Handlers) handleContractQuery(w http.ResponseWriter, r *http.Request) {
	contract, err, status := cdigest.ParseRequest(w, r, "contract")
	if err != nil {
		cdigest.HTTP2ProblemWithError(w, err, status)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxContractQueryBodyBytes)
	body := &bytes.Buffer{}
	if _, err := body.ReadFrom(r.Body); err != nil {
		var maxBytesErr *http.MaxBytesError
		if stderrors.As(err, &maxBytesErr) {
			cdigest.HTTP2ProblemWithError(
				w,
				pkgerrors.Errorf("query body exceeds max size: max %d bytes", MaxContractQueryBodyBytes),
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		cdigest.HTTP2ProblemWithError(w, err, http.StatusInternalServerError)
		return
	}

	var callData map[string]string
	if err := json.Unmarshal(body.Bytes(), &callData); err != nil {
		cdigest.HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	}
	if err := runtime.ValidateContractCallDataLimits("query callData", callData); err != nil {
		cdigest.HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	}

	fName, found := callData["function"]
	if !found || fName == "" {
		cdigest.HTTP2ProblemWithError(w, pkgerrors.Errorf("missing function in query body"), http.StatusBadRequest)
		return
	}

	b, err := hd.handleContractQueryInGroup(contract, callData)
	if err != nil {
		cdigest.HTTP2HandleError(w, err)
		return
	}

	cdigest.HTTP2WriteHalBytes(hd.encoder, w, b, http.StatusOK)
}

func (hd *Handlers) handleContractQueryInGroup(contract string, callData map[string]string) ([]byte, error) {
	contractAddr, design, designState, err := ContractDesignFromChainState(hd.database, contract)
	if err != nil {
		return nil, err
	}
	designStateValue, err := state.GetDesignStateValueFromState(designState)
	if err != nil {
		return nil, err
	}

	queryHeight := designState.Height()
	currentHeight := hd.database.LastBlock()
	responseState := designState

	_, runtimeValue, _, runtimeFound, err := ContractRuntimeFromChainState(hd.database, contract)
	if err != nil {
		return nil, err
	}
	if runtimeFound && runtimeValue.Engine == state.RuntimeEngineGnoSnapshot {
		_, _, snapshotState, snapshotFound, err := ContractSnapshotFromChainState(hd.database, contract)
		if err != nil {
			return nil, err
		}
		if !snapshotFound {
			return nil, pkgerrors.Errorf("snapshot state not found for typed contract %s", contract)
		}

		// For snapshot-backed Gno contracts, snapshot state height is the canonical query height.
		queryHeight = snapshotState.Height()
		responseState = snapshotState
	}
	if currentHeight < queryHeight {
		currentHeight = queryHeight
	}

	var schema *runtime.ContractSchema
	if persistedSchema, ok := runtime.RuntimeSchemaFromPersisted(design.ContractCode(), designStateValue.Schema); ok {
		schema = &persistedSchema
	}

	qr, qerr := digestContractQueryEngine.QueryContract(
		*hd.encoders,
		hd.database.State,
		runtime.QueryRequest{
			Contract:      contractAddr,
			Sender:        contractAddr,
			Height:        queryHeight,
			CurrentHeight: currentHeight,
			ContractCode:  design.ContractCode(),
			Schema:        schema,
			Function:      callData["function"],
			CallData:      callData,
		},
	)
	if qerr != nil {
		return nil, pkgerrors.Errorf("%v", qerr)
	}

	i, err := hd.buildContractQuery(contract, callData["function"], qr, responseState)
	if err != nil {
		return nil, err
	}

	return hd.encoder.Marshal(i)
}

func (hd *Handlers) buildContractQuery(
	contract string,
	function string,
	qr runtime.QueryResult,
	st base.State,
) (cdigest.Hal, error) {
	h, err := hd.combineURL(HandlerPathContractQuery, "contract", contract)
	if err != nil {
		return nil, err
	}

	resp := ContractQueryResponse{
		Contract: contract,
		Function: function,
		Engine:   string(qr.Engine),
		ReadOnly: true,
		Output: ContractQueryOutput{
			Result: qr.Result,
			Ok:     qr.Ok,
		},
	}

	var hal cdigest.Hal
	hal = cdigest.NewBaseHal(resp, cdigest.NewHalLink(h, nil))

	h, err = hd.combineURL(HandlerPathContractDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("design", cdigest.NewHalLink(h, nil))

	h, err = hd.combineURL(cdigest.HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", cdigest.NewHalLink(h, nil))

	return hal, nil
}
