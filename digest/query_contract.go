package digest

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"net/http"

	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	pkgerrors "github.com/pkg/errors"
)

var digestContractQueryEngine cruntime.ContractEngine = cruntime.NewGnoEngine()

const MaxContractQueryBodyBytes = 128 * 1024

type ContractQueryResponse struct {
	Contract string      `json:"contract"`
	Function string      `json:"function"`
	Engine   string      `json:"engine"`
	Result   interface{} `json:"result"`
	Ok       *bool       `json:"ok,omitempty"`
	ReadOnly bool        `json:"read_only"`
}

func (hd *Handlers) handleContractQuery(w http.ResponseWriter, r *http.Request) {
	contract, err, status := ParseRequest(w, r, "contract")
	if err != nil {
		HTTP2ProblemWithError(w, err, status)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxContractQueryBodyBytes)
	body := &bytes.Buffer{}
	if _, err := body.ReadFrom(r.Body); err != nil {
		var maxBytesErr *http.MaxBytesError
		if stderrors.As(err, &maxBytesErr) {
			HTTP2ProblemWithError(
				w,
				pkgerrors.Errorf("query body exceeds max size: max %d bytes", MaxContractQueryBodyBytes),
				http.StatusRequestEntityTooLarge,
			)
			return
		}

		HTTP2ProblemWithError(w, err, http.StatusInternalServerError)
		return
	}

	var callData map[string]string
	if err := json.Unmarshal(body.Bytes(), &callData); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	}
	if err := cruntime.ValidateContractCallDataLimits("query callData", callData); err != nil {
		HTTP2ProblemWithError(w, err, http.StatusBadRequest)
		return
	}

	fName, found := callData["function"]
	if !found || fName == "" {
		HTTP2ProblemWithError(w, pkgerrors.Errorf("missing function in query body"), http.StatusBadRequest)
		return
	}

	b, err := hd.handleContractQueryInGroup(contract, callData)
	if err != nil {
		HTTP2HandleError(w, err)
		return
	}

	HTTP2WriteHalBytes(hd.enc, w, b, http.StatusOK)
}

func (hd *Handlers) handleContractQueryInGroup(contract string, callData map[string]string) ([]byte, error) {
	contractAddr, design, designState, err := ContractDesignFromChainState(hd.database, contract)
	if err != nil {
		return nil, err
	}

	queryHeight := designState.Height()
	responseState := designState

	_, runtimeValue, _, runtimeFound, err := ContractRuntimeFromChainState(hd.database, contract)
	if err != nil {
		return nil, err
	}
	if runtimeFound && runtimeValue.Engine == pstate.RuntimeEngineGnoSnapshot {
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

	sender, err := cruntime.ParseOptionalQuerySender(*hd.encs, contractAddr, callData)
	if err != nil {
		return nil, err
	}

	qr, qerr := digestContractQueryEngine.QueryContract(
		*hd.encs,
		hd.database.State,
		cruntime.QueryRequest{
			Contract:     contractAddr,
			Sender:       sender,
			Height:       queryHeight,
			ContractCode: design.ContractCode(),
			Function:     callData["function"],
			CallData:     callData,
		},
	)
	if qerr != nil {
		return nil, pkgerrors.Errorf("%v", qerr)
	}

	i, err := hd.buildContractQuery(contract, callData["function"], qr, responseState)
	if err != nil {
		return nil, err
	}

	return hd.enc.Marshal(i)
}

func (hd *Handlers) buildContractQuery(
	contract string,
	function string,
	qr cruntime.QueryResult,
	st base.State,
) (Hal, error) {
	h, err := hd.combineURL(HandlerPathContractQuery, "contract", contract)
	if err != nil {
		return nil, err
	}

	resp := ContractQueryResponse{
		Contract: contract,
		Function: function,
		Engine:   string(qr.Engine),
		Result:   qr.Result,
		Ok:       qr.Ok,
		ReadOnly: true,
	}

	var hal Hal
	hal = NewBaseHal(resp, NewHalLink(h, nil))

	h, err = hd.combineURL(HandlerPathContractDesign, "contract", contract)
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("design", NewHalLink(h, nil))

	h, err = hd.combineURL(HandlerPathBlockByHeight, "height", st.Height().String())
	if err != nil {
		return nil, err
	}
	hal = hal.AddLink("block", NewHalLink(h, nil))

	return hal, nil
}
