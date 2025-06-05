package node

import (
	"fmt"
	"sort"

	"github.com/x3pi/Hashgraph/event"
	// You'll need to adjust the import path based on your Go module name
	// e.g., "your_module_name/event"
)

// Node represents a participant in the Hashgraph network.
type Node struct {
	ID       int
	Events   map[string]event.Event // Stores events known to this node (Hash -> Event)
	Balances map[string]int         // Account balances known to this node
}

// NewNode creates a new node with a given ID and initial balances.
func NewNode(id int, initialBalances map[string]int) *Node {
	// Create a copy of initialBalances to avoid modification of the original map
	balancesCopy := make(map[string]int)
	for k, v := range initialBalances {
		balancesCopy[k] = v
	}
	return &Node{
		ID:       id,
		Events:   make(map[string]event.Event),
		Balances: balancesCopy,
	}
}

// AddEvent adds an event to the node's local store if not already present.
func (n *Node) AddEvent(e event.Event) {
	if _, exists := n.Events[e.Hash]; !exists {
		n.Events[e.Hash] = e
		// fmt.Printf("ℹ️ Node %d received event %s\n", n.ID, e.Hash[:6])
	}
}

// Gossip simulates the gossiping of events to peers.
// Each node sends events it knows to other peers.
func (n *Node) Gossip(peers []*Node) {
	for _, peer := range peers {
		if peer.ID == n.ID {
			continue // Don't gossip to self
		}
		for _, eventToSend := range n.Events {
			// Check if peer already has this event by hash
			if _, exists := peer.Events[eventToSend.Hash]; !exists {
				// To simulate network transfer, the peer receives a copy
				receivedEvent := eventToSend
				// If votes are part of the event, they are gossiped too.
				// Ensure the map is copied if it's to be modified independently by the peer,
				// but here we are copying the whole event struct.
				peer.Events[receivedEvent.Hash] = receivedEvent
				fmt.Printf("📢 Node %d gửi giao dịch %s đến Node %d\n", n.ID, receivedEvent.Hash[:6], peer.ID)
			}
		}
	}
}

// VirtualVoting simulates the virtual voting process for events known to this node.
// In this simplified version, a node votes for events it knows about.
// If an event accumulates enough votes (2/3+1), it's considered valid by this node.
func (n *Node) VirtualVoting(totalNodes int) {
	// Threshold for consensus (supermajority, e.g., > 2/3)
	// For N nodes, threshold is floor(2N/3) + 1.
	// Or simply, if count > 2N/3, which means count >= floor(2N/3) + 1.
	// The original code used `countVotes >= threshold` with `threshold := (2 * totalNodes) / 3`.
	// This means if totalNodes = 4, threshold = (2*4)/3 = 8/3 = 2. `countVotes >= 2`.
	// For N=4, 2/3 * N = 2.66. Supermajority needs 3 votes.
	// Let's adjust to common BFT: threshold = (2*totalNodes)/3 + 1, if integer arithmetic is tricky,
	// it's often easier to check `countVotes * 3 > 2 * totalNodes`.
	threshold := (2*totalNodes)/3 + 1
	if totalNodes <= 0 { // avoid division by zero or negative totalNodes
		return
	}

	updatedEvents := make(map[string]event.Event) // To store modifications

	for hash, currentEvent := range n.Events {
		// A node implicitly "votes" for all events it has seen and hasn't found to be invalid.
		// Here, we simply mark that this node has seen/processed it.
		// The original logic: event.Votes[n.ID] = true, then count all true votes.
		if currentEvent.Votes == nil {
			currentEvent.Votes = make(map[int]bool)
		}
		currentEvent.Votes[n.ID] = true // This node votes for this event

		countVotes := 0
		for _, voted := range currentEvent.Votes {
			if voted {
				countVotes++
			}
		}

		// If the event hasn't reached consensus yet, check if it does now
		if !currentEvent.Valid && countVotes >= threshold {
			currentEvent.Valid = true
			fmt.Printf("👍 Node %d: Giao dịch %s đạt đồng thuận với %d/%d phiếu.\n", n.ID, currentEvent.Hash[:6], countVotes, totalNodes)
		}
		updatedEvents[hash] = currentEvent
	}
	n.Events = updatedEvents // Apply all updates
}

// ExecuteTransactions processes events that are marked as valid and have a consensus timestamp.
// Transactions are executed in order of their consensus timestamp.
func (n *Node) ExecuteTransactions() {
	fmt.Printf("\n⏳ Node %d thực thi giao dịch:\n", n.ID)

	// Filter for valid events with a consensus time
	var transactionsToExecute []event.Event
	for _, e := range n.Events {
		if e.Valid && e.ConsensusTime > 0 { // Ensure consensus time is set
			transactionsToExecute = append(transactionsToExecute, e)
		}
	}

	if len(transactionsToExecute) == 0 {
		fmt.Println("   (Không có giao dịch hợp lệ hoặc đã được đồng thuận về thời gian để thực thi)")
		return
	}

	// Sort transactions: primary by ConsensusTime, secondary by Hash (for determinism)
	sort.Slice(transactionsToExecute, func(i, j int) bool {
		if transactionsToExecute[i].ConsensusTime == transactionsToExecute[j].ConsensusTime {
			return transactionsToExecute[i].Hash < transactionsToExecute[j].Hash
		}
		return transactionsToExecute[i].ConsensusTime < transactionsToExecute[j].ConsensusTime
	})

	for i, e := range transactionsToExecute {
		// Check sender's balance
		if e.Sender == "" { // Handle minting or events without a sender if applicable
			n.Balances[e.Receiver] += e.Amount
			fmt.Printf("✔ [%d] %s: Coinbase/Mint → %s (💰 %d coin)\n",
				i+1, e.Hash[:6], e.Receiver, e.Amount)
			continue
		}

		if n.Balances[e.Sender] >= e.Amount {
			n.Balances[e.Sender] -= e.Amount
			n.Balances[e.Receiver] += e.Amount
			fmt.Printf("✔ [%d] %s: %s → %s (💰 %d coin)\n",
				i+1, e.Hash[:6], e.Sender, e.Receiver, e.Amount)
		} else {
			// This event was deemed valid by consensus, but the sender might not have funds
			// in this node's view if balances aren't perfectly synced or if this is an optimistic execution.
			// For simplicity, we just reject. In a real system, this would be more complex.
			fmt.Printf("❌ bị từ chối do không đủ số dư [%d] %s: %s → %s (💰 %d coin) (Số dư hiện tại của %s: %d)\n",
				i+1, e.Hash[:6], e.Sender, e.Receiver, e.Amount, e.Sender, n.Balances[e.Sender])
		}
	}
}

// PrintBalances prints the current balances known to the node.
func (n *Node) PrintBalances() {
	fmt.Printf("💰 Node %d Balances:\n", n.ID)
	// Sort keys for consistent output
	var accounts []string
	for acc := range n.Balances {
		accounts = append(accounts, acc)
	}
	sort.Strings(accounts)
	for _, acc := range accounts {
		fmt.Printf("   - %s: %d\n", acc, n.Balances[acc])
	}
}

// PrintEvents prints the events known to the node.
func (n *Node) PrintEvents(printDetails bool) {
	fmt.Printf("📄 Node %d Events (%d total):\n", n.ID, len(n.Events))
	if !printDetails || len(n.Events) == 0 {
		return
	}

	// Sort events by hash for consistent output
	var eventHashes []string
	for hash := range n.Events {
		eventHashes = append(eventHashes, hash)
	}
	sort.Strings(eventHashes)

	for _, hash := range eventHashes {
		e := n.Events[hash]
		fmt.Printf("   - Hash: %s... (Sender: %s, Receiver: %s, Amount: %d, Valid: %t, Votes: %d, CT: %d)\n",
			e.Hash[:6], e.Sender, e.Receiver, e.Amount, e.Valid, len(e.Votes), e.ConsensusTime)
	}
}
