package digest

import (
	"bytes"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/operation/contract/runtime"
	"github.com/imfact-labs/smart-contract-model/state"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCanonicalUint256EquivalentSnapshotsHaveSameDigestIdentity(t *testing.T) {
	_, enc := newDigestTestEncoders(t)
	snapshot := []byte(`{"version":1,"bindings":[{"name":"totalSupply","value":{"kind":"scalar","scalar":"42"}}]}`)
	values := []state.SnapshotStateValue{
		state.NewSnapshotStateValue(runtime.GnoSnapshotVersion, runtime.GnoSnapshotCodecName, append([]byte(nil), snapshot...)),
		state.NewSnapshotStateValue(runtime.GnoSnapshotVersion, runtime.GnoSnapshotCodecName, append([]byte(nil), snapshot...)),
	}
	var hashes [][]byte
	var digestHashes []string
	for i, value := range values {
		st := common.NewBaseState(base.Height(i+1), state.SnapshotStateKey(base.NewStringAddress("contractu256digest")), value, nil, nil)
		doc, err := NewContractSnapshotDoc(st, enc)
		if err != nil {
			t.Fatalf("NewContractSnapshotDoc: %v", err)
		}
		body, err := doc.MarshalBSON()
		if err != nil {
			t.Fatalf("MarshalBSON: %v", err)
		}
		var decoded bson.M
		if err := bson.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode digest BSON: %v", err)
		}
		hash, ok := decoded["snapshot_sha256"].(string)
		if !ok || hash == "" {
			t.Fatalf("snapshot_sha256 missing: %#v", decoded)
		}
		hashes = append(hashes, value.HashBytes())
		digestHashes = append(digestHashes, hash)
	}
	if !bytes.Equal(hashes[0], hashes[1]) {
		t.Fatal("equivalent canonical snapshots have different SnapshotStateValue hashes")
	}
	if digestHashes[0] != digestHashes[1] {
		t.Fatalf("equivalent canonical snapshots have different snapshot_sha256: %q != %q", digestHashes[0], digestHashes[1])
	}
}
