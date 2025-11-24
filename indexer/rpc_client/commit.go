package rpcclient

import (
	"fmt"
	"strconv"
	"time"
)

// CommitResponse is the response from the commit endpoint
type CommitResponse struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Result  CommitResult  `json:"result"`
}

// CommitResult is the result from the commit endpoint
type CommitResult struct {
	SignedHeader SignedHeader `json:"signed_header"`
}

// SignedHeader is the struct for the signed header, which is a part of commit result
type SignedHeader struct {
	Header BlockHeader `json:"header"`
	Commit Commit      `json:"commit"`
}

// Commit is the struct for the commit, which is a part of signed header
type Commit struct {
	BlockID    BlockID      `json:"block_id"`
	Precommits []*Precommit `json:"precommits"`
}

// GetHeight returns the height of the commit
func (cr *CommitResponse) GetHeight() (uint64, error) {
	if cr == nil {
		return 0, fmt.Errorf("CommitResponse Header Height is nil")
	}
	return strconv.ParseUint(cr.Result.SignedHeader.Header.Height, 10, 64)
}

// GetTimestamp returns the timestamp of the commit
func (cr *CommitResponse) GetTimestamp() time.Time {
	if cr == nil {
		return time.Time{}
	}
	return cr.Result.SignedHeader.Header.Time
}

// GetSigners returns the signers of the commit
func (cr *CommitResponse) GetSigners() []*Precommit {
	if cr == nil {
		return nil
	}
	return cr.Result.SignedHeader.Commit.Precommits
}

// GetProposerAddress returns the proposer address of the commit
func (cr *CommitResponse) GetProposerAddress() string {
	if cr == nil {
		return ""
	}
	return cr.Result.SignedHeader.Header.ProposerAddress
}
