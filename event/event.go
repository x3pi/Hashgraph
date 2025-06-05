package event

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Event represents a transaction or event in the Hashgraph.
type Event struct {
	Sender        string
	Receiver      string
	Amount        int
	Parents       []string     // Hashes of parent events
	Timestamp     int64        // Creation timestamp
	ConsensusTime int64        // Timestamp determined by consensus
	Hash          string       // Hash of this event
	Votes         map[int]bool // Votes received from nodes (NodeID -> true)
	Valid         bool         // Whether the event has reached consensus and is valid
	RoundReceived int          // For more advanced consensus, round event was received by a node (not used in this simplified version yet)
	IsWitness     bool         // For more advanced consensus, if this event is a witness (not used in this simplified version yet)
}

// NewEvent creates and returns a new event.
// The hash is calculated based on sender, receiver, amount, and timestamp.
func NewEvent(sender, receiver string, amount int, parents []string) Event {
	timestamp := time.Now().UnixNano()
	// Simple hash input string; in a real scenario, ensure canonical representation
	hashInput := sender + receiver + fmt.Sprint(amount) + fmt.Sprint(timestamp)
	for _, p := range parents {
		hashInput += p
	}
	hashBytes := sha256.Sum256([]byte(hashInput))
	eventHash := hex.EncodeToString(hashBytes[:])

	return Event{
		Sender:        sender,
		Receiver:      receiver,
		Amount:        amount,
		Parents:       parents,
		Timestamp:     timestamp,
		ConsensusTime: 0, // Initialized to 0, will be set by consensus algorithm
		Hash:          eventHash,
		Votes:         make(map[int]bool),
		Valid:         false, // Initialized to false
	}
}
