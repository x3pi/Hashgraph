package main

import (
	"fmt"
	"math/rand"
	"sort" // Thêm import này để sắp xếp
	"time"

	"github.com/x3pi/Hashgraph/consensus"
	"github.com/x3pi/Hashgraph/event"
	"github.com/x3pi/Hashgraph/node"
	// Adjust import paths according to your Go module structure
	// e.g., "your_module_name/consensus", "your_module_name/event", "your_module_name/node"
)

const (
	numNodes                 = 4
	numSimulationRounds      = 10 // Main simulation rounds
	gossipIterationsPerRound = 2  // How many times gossip runs per simulation round
	votingIterationsPerRound = 2  // How many times virtual voting runs per simulation round
)

func main() {
	rand.Seed(time.Now().UnixNano()) // Initialize random seed

	// Initial balances
	initialBalances := map[string]int{
		"Alice":   1000,
		"Bob":     500,
		"Charlie": 750,
		"David":   600,
		"System":  0, // For potential fees or new coin generation
	}

	// Initialize nodes
	nodes := make([]*node.Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = node.NewNode(i+1, initialBalances) // Node IDs start from 1
	}

	fmt.Println("--- Khởi tạo mạng Hashgraph ---")
	for _, n := range nodes {
		n.PrintBalances()
	}
	fmt.Println("---------------------------------")

	// Create some initial events and distribute them
	// Parents are empty for these initial events
	event1 := event.NewEvent("Alice", "Bob", 30, []string{})
	event2 := event.NewEvent("Bob", "Charlie", 20, []string{})
	event3 := event.NewEvent("Alice", "David", 800, []string{}) // Potentially not enough funds
	event4 := event.NewEvent("Charlie", "Alice", 10, []string{})

	// Distribute initial events to different nodes to simulate them originating from those nodes
	if len(nodes) > 0 {
		nodes[0].AddEvent(event1)
		fmt.Printf("🎉 Node %d tạo giao dịch ban đầu: %s (%s -> %s, %d coin)\n", nodes[0].ID, event1.Hash[:6], event1.Sender, event1.Receiver, event1.Amount)
	}
	if len(nodes) > 1 {
		nodes[1].AddEvent(event2)
		fmt.Printf("🎉 Node %d tạo giao dịch ban đầu: %s (%s -> %s, %d coin)\n", nodes[1].ID, event2.Hash[:6], event2.Sender, event2.Receiver, event2.Amount)
	}
	if len(nodes) > 2 {
		nodes[2].AddEvent(event3)
		fmt.Printf("🎉 Node %d tạo giao dịch ban đầu: %s (%s -> %s, %d coin)\n", nodes[2].ID, event3.Hash[:6], event3.Sender, event3.Receiver, event3.Amount)
	}
	if len(nodes) > 3 {
		nodes[3].AddEvent(event4)
		fmt.Printf("🎉 Node %d tạo giao dịch ban đầu: %s (%s -> %s, %d coin)\n", nodes[3].ID, event4.Hash[:6], event4.Sender, event4.Receiver, event4.Amount)
	}

	// --- Main Simulation Loop ---
	fmt.Printf("\n--- Bắt đầu %d vòng mô phỏng ---\n", numSimulationRounds)
	for r := 0; r < numSimulationRounds; r++ {
		fmt.Printf("\n🔄 Vòng %d/%d 🔄\n", r+1, numSimulationRounds)

		// =================================================================
		// ===== TẠO GIAO DỊCH MỚI TRONG MỖI VÒNG LẶP (TRỪ VÒNG ĐẦU) =====
		// =================================================================
		if r > 0 {
			creatorNodeIndex := rand.Intn(numNodes)
			creatorNode := nodes[creatorNodeIndex]

			accounts := []string{"Alice", "Bob", "Charlie", "David"}
			senderIndex := rand.Intn(len(accounts))
			receiverIndex := rand.Intn(len(accounts))
			for senderIndex == receiverIndex {
				receiverIndex = rand.Intn(len(accounts))
			}
			sender := accounts[senderIndex]
			receiver := accounts[receiverIndex]
			amount := rand.Intn(50) + 1

			var parentHashes []string
			// (Logic chọn parent phức tạp hơn có thể được thêm vào đây nếu cần)

			newEvent := event.NewEvent(sender, receiver, amount, parentHashes)
			creatorNode.AddEvent(newEvent)
			fmt.Printf("✨ Node %d tạo giao dịch mới trong Vòng %d: %s (%s -> %s, %d coin)\n",
				creatorNode.ID, r+1, newEvent.Hash[:6], newEvent.Sender, newEvent.Receiver, newEvent.Amount)
		}
		// =================================================================
		// ===== KẾT THÚC TẠO GIAO DỊCH MỚI ===============================
		// =================================================================

		// --- Gossip Phase ---
		for i := 0; i < gossipIterationsPerRound; i++ {
			for _, n := range nodes {
				n.Gossip(nodes)
			}
		}

		// --- Virtual Voting Phase ---
		for i := 0; i < votingIterationsPerRound; i++ {
			for _, n := range nodes {
				n.VirtualVoting(numNodes)
			}
		}

		// =====================================================================
		// ===== TÍNH TOÁN THỜI GIAN ĐỒNG THUẬN TOÀN CỤC CHO VÒNG NÀY =====
		// =====================================================================
		fmt.Printf("--- Vòng %d: Tính toán Thời gian Đồng thuận Toàn cục ---\n", r+1)
		consensus.CalculateGlobalConsensusTime(nodes)
		// =====================================================================
		// ===== KẾT THÚC TÍNH TOÁN THỜI GIAN ĐỒNG THUẬN =======================
		// =====================================================================

		// =====================================================================
		// ===== THỰC THI GIAO DỊCH CHO VÒNG NÀY =====
		// =====================================================================
		fmt.Printf("--- Vòng %d: Thực thi Giao dịch ---\n", r+1)
		for _, n := range nodes {
			n.ExecuteTransactions()
		}
		// =====================================================================
		// ===== KẾT THÚC THỰC THI GIAO DỊCH ===================================
		// =====================================================================

		// =====================================================================
		// ===== IN TRẠNG THÁI ĐỒNG THUẬN CHI TIẾT SAU MỖI VÒNG =====
		// =====================================================================
		fmt.Printf("--- Vòng %d: Trạng thái Đồng thuận Chi tiết (Sau thực thi) ---\n", r+1)
		for _, n := range nodes {
			fmt.Printf("  NODE %d (Tổng số sự kiện: %d):\n", n.ID, len(n.Events))
			if len(n.Events) == 0 {
				fmt.Println("    Không có sự kiện nào.")
			}
			var eventHashes []string
			for hash := range n.Events {
				eventHashes = append(eventHashes, hash)
			}
			sort.Strings(eventHashes)

			for _, hash := range eventHashes {
				event := n.Events[hash]
				trueVotes := 0
				for _, vote := range event.Votes {
					if vote {
						trueVotes++
					}
				}
				fmt.Printf("    - Sự kiện %s: Hợp lệ (Valid): %t, Số phiếu (True Votes): %d/%d (CT: %d, Tạo lúc: %d)\n",
					event.Hash[:6], event.Valid, trueVotes, len(event.Votes), event.ConsensusTime, event.Timestamp)
			}
		}
		// =====================================================================
		// ===== KẾT THÚC IN TRẠNG THÁI ĐỒNG THUẬN =================================
		// =====================================================================

		// =====================================================================
		// ===== IN TRẠNG THÁI BALANCE SAU MỖI VÒNG (SAU THỰC THI) =====
		// =====================================================================
		fmt.Printf("--- Vòng %d: Trạng thái Số dư (Balances) (Sau thực thi) ---\n", r+1)
		for _, n := range nodes {
			n.PrintBalances()
		}
		fmt.Println("---------------------------------------------")
		// =====================================================================
		// ===== KẾT THÚC IN TRẠNG THÁI BALANCE ===============================
		// =====================================================================

	}
	fmt.Println("\n--- Kết thúc các vòng mô phỏng ---")

	// Không cần gọi lại CalculateGlobalConsensusTime và ExecuteTransactions ở đây nữa
	// vì chúng đã được thực hiện trong mỗi vòng.

	// --- Final State ---
	fmt.Println("\n--- Trạng thái cuối cùng của mạng (Sau tất cả các vòng và thực thi) ---")
	for _, n := range nodes {
		n.PrintBalances()
		n.PrintEvents(false) // Set to true for detailed event list
		fmt.Println("---")
	}

	fmt.Println("\n--- Chi tiết sự kiện cuối cùng trên Node 1 (Sau tất cả các vòng) ---")
	if len(nodes) > 0 {
		nodes[0].PrintEvents(true)
	}
}
