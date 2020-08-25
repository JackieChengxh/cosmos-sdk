package exported

import (
	"encoding/json"
	"fmt"

	ics23 "github.com/confio/ics23/go"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ClientState defines the required common functions for light clients.
type ClientState interface {
	GetChainID() string
	ClientType() ClientType
	GetLatestHeight() Height
	IsFrozen() bool
	GetFrozenHeight() uint64
	Validate() error
	GetProofSpecs() []*ics23.ProofSpec

	// Update and Misbehaviour functions

	CheckHeaderAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Header) (ClientState, ConsensusState, error)
	CheckMisbehaviourAndUpdateState(sdk.Context, codec.BinaryMarshaler, sdk.KVStore, Misbehaviour) (ClientState, error)

	// State verification functions

	VerifyClientState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root commitmentexported.Root,
		height uint64,
		prefix commitmentexported.Prefix,
		counterpartyClientIdentifier string,
		proof []byte,
		clientState ClientState,
	) error
	VerifyClientConsensusState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		root commitmentexported.Root,
		height Height,
		counterpartyClientIdentifier string,
		consensusHeight Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		consensusState ConsensusState,
	) error
	VerifyConnectionState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		connectionID string,
		connectionEnd connectionexported.ConnectionI,
	) error
	VerifyChannelState(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		channel channelexported.ChannelI,
	) error
	VerifyPacketCommitment(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		commitmentBytes []byte,
	) error
	VerifyPacketAcknowledgement(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
		acknowledgement []byte,
	) error
	VerifyPacketAcknowledgementAbsence(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		sequence uint64,
	) error
	VerifyNextSequenceRecv(
		store sdk.KVStore,
		cdc codec.BinaryMarshaler,
		height Height,
		prefix commitmentexported.Prefix,
		proof []byte,
		portID,
		channelID string,
		nextSequenceRecv uint64,
	) error
}

// ConsensusState is the state of the consensus process
type ConsensusState interface {
	ClientType() ClientType // Consensus kind

	// GetHeight returns the height of the consensus state
	GetHeight() Height

	// GetRoot returns the commitment root of the consensus state,
	// which is used for key-value pair verification.
	GetRoot() commitmentexported.Root

	// GetTimestamp returns the timestamp (in nanoseconds) of the consensus state
	GetTimestamp() uint64

	ValidateBasic() error
}

// Misbehaviour defines a specific consensus kind and an evidence
type Misbehaviour interface {
	evidenceexported.Evidence
	GetIBCHeight() Height
	ClientType() ClientType
	GetClientID() string
}

// Header is the consensus state update information
type Header interface {
	ClientType() ClientType
	GetHeight() Height
}

// message types for the IBC client
const (
	TypeMsgCreateClient             string = "create_client"
	TypeMsgUpdateClient             string = "update_client"
	TypeMsgSubmitClientMisbehaviour string = "submit_client_misbehaviour"
)

// MsgCreateClient defines the msg interface that the
// CreateClient Handler expects
type MsgCreateClient interface {
	sdk.Msg
	GetClientID() string
	GetClientType() string
	GetConsensusState() ConsensusState
	InitializeClientState() ClientState
}

// MsgUpdateClient defines the msg interface that the
// UpdateClient Handler expects
type MsgUpdateClient interface {
	sdk.Msg
	GetClientID() string
	GetHeader() Header
}

// ClientType defines the type of the consensus algorithm
type ClientType byte

// available client types
const (
	SoloMachine ClientType = 6
	Tendermint  ClientType = 7
	Localhost   ClientType = 9
)

// string representation of the client types
const (
	ClientTypeSoloMachine string = "solomachine"
	ClientTypeTendermint  string = "tendermint"
	ClientTypeLocalHost   string = "localhost"
)

func (ct ClientType) String() string {
	switch ct {
	case Tendermint:
		return ClientTypeTendermint
	case Localhost:
		return ClientTypeLocalHost
	default:
		return ""
	}
}

// MarshalJSON marshal to JSON using string.
func (ct ClientType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

// UnmarshalJSON decodes from JSON.
func (ct *ClientType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	clientType := ClientTypeFromString(s)
	if clientType == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "invalid client type '%s'", s)
	}

	*ct = clientType
	return nil
}

// ClientTypeFromString returns a byte that corresponds to the registered client
// type. It returns 0 if the type is not found/registered.
func ClientTypeFromString(clientType string) ClientType {
	switch clientType {
	case ClientTypeTendermint:
		return Tendermint
	case ClientTypeLocalHost:
		return Localhost
	default:
		return 0
	}
}

// Height is a monotonically increasing data type
// that can be compared against another Height for the purposes of updating and freezing clients
//
// Normally the EpochHeight is incremented at each height while keeping epoch number the same
// However some consensus algorithms may choose to reset the height in certain conditions
// e.g. hard forks, state-machine breaking changes
// In these cases, the epoch number is incremented so that height continues to be monitonically increasing
// even as the EpochHeight gets reset
type Height struct {
	EpochNumber uint64 `json:"epoch_number" yaml:"epoch_number"` // The current epoch of the chain
	EpochHeight uint64 `json:"epoch_height" yaml:"epoch_height"` // The height of the chain at the given epoch
}

// NewHeight creates a new client Height instance from the epoch values.
func NewHeight(epochNumber, epochHeight uint64) Height {
	return Height{
		EpochNumber: epochNumber,
		EpochHeight: epochHeight,
	}
}

/// Compare implements a method to compare two heights. When comparing two heights a, b
// we can call a.Compare(b) which will return
// -1 if a < b
// 0  if a = b
// 1  if a > b
//
// It first compares based on epoch numbers, whichever has the higher epoch number is the higher height
// If epoch number is the same, then the epoch height is compared
func (h Height) Compare(other Height) int64 {
	var cmp int64
	if h.EpochNumber != other.EpochNumber {
		cmp = int64(h.EpochNumber) - int64(other.EpochNumber)
	} else {
		cmp = int64(h.EpochHeight) - int64(other.EpochHeight)
	}
	if cmp < 0 {
		return -1
	} else if cmp > 0 {
		return 1
	}
	return 0
}

// LT Helper comparison function returns true if h < other
func (h Height) LT(other Height) bool {
	return h.Compare(other) == -1
}

// GT Helper comparison function returns true if h > other
func (h Height) GT(other Height) bool {
	return h.Compare(other) == 1
}

// EQ Helper comparison function returns true if h == other
func (h Height) EQ(other Height) bool {
	return h.Compare(other) == 0
}

// Decrement will return a decremented height from the given height. If this is not possible,
// an error is returned
// Decrement will return a new height with the EpochHeight decremented
// If the EpochHeight is already at lowest value (1), then false success flag is returend
func (h Height) Decrement() (decremented Height, success bool) {
	if h.EpochHeight <= 1 {
		return Height{}, false
	}
	return NewHeight(h.EpochNumber, h.EpochHeight-1), true
}

// Valid returns true if height is valid, false otherwise
// Returns false if EpochHeight is 0
func (h Height) Valid() bool {
	return h.EpochHeight != 0
}

// String implements the fmt.Stringer interface
func (h Height) String() string {
	return fmt.Sprintf("epoch-%d_height-%d", h.EpochNumber, h.EpochHeight)
}

// IsZero returns true if height epoch and epoch-height are both 0
func (h Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
