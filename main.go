package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main(){
	// Flags
	idPtr := flag.Int("id", 0, "ID númerico do aluno") 
	peerIP := flag.String("peer", "", "IP do colega para entrar na rede. ex: -peer 192.168.1.10")
	flag.Parse()

	if *idPtr <= 0 {
		printError("Erro: Você precisa informar um ID de aluno válido e maior que zero.")
		printError("Use o parâmetro -id. Exemplo: go run . -id 5")
		os.Exit(1)
	}

	localIP = getLocalIP()
	printInfo("Seu IP é: %s", localIP)

	peerID := fmt.Sprintf("ALUNO-%02d", *idPtr)
	stickerID := fmt.Sprintf("FIG-%02d", *idPtr)
	inventoryFile = fmt.Sprintf("inventory-%s.json", peerID)
	
	// Peer node
	node = &Peer{
		ID:          peerID,
		StickerID:   stickerID,
		Inventory:   map[string]int{stickerID: 28},
		Neighbors:   make(map[string]*PeerConn),
		KnownPeers:  []string{},
		SeenQueries: make(map[string]time.Time),
		SearchResults: make(map[string]*PeerConn),
	}

	if *peerIP != "" {
		addr := fmt.Sprintf("ws://%s:8080/ws", *peerIP)
		go connectToPeer(addr)
	}

	initializeInventory(*idPtr)

	// Web Socket
	http.HandleFunc("/ws", wsHandler)
	registerAPIRoutes()
	printSystem("Nó %s iniciado na porta 8080", peerID)

	go inputLoop()	

	go func() {
		for range time.Tick(1 * time.Minute) {
			node.mu.Lock()
			for id, t := range node.SeenQueries {
				if time.Since(t) > 5*time.Minute {
					delete(node.SeenQueries, id)
				}
			}
			node.mu.Unlock()
		}
	}()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		printError("Erro ao iniciar servidor: %v", err)
	}
}

func inputLoop() {
    scanner := bufio.NewScanner(os.Stdin)
    printMenu()

    for {
        if !scanner.Scan() {
            break
        }
        line := strings.TrimSpace(scanner.Text())
        parts := strings.Fields(line)
        if len(parts) == 0 {
            continue
        }

        switch parts[0] {
        case "search":
            if len(parts) < 2 {
                printWarning("Uso: search <FIG-XX>")
				fmt.Print("> ")
                continue
            }
			if err := cmdSearch(parts[1]); err != nil {
				printWarning("%v", err)
				fmt.Print("> ")
				continue
			}

		case "offer":
			if len(parts) < 2 {
				printWarning("Uso: offer <FIG-XX>")
				fmt.Print("> ")
				continue
			}
			if err := cmdOffer(parts[1]); err != nil {
				printWarning("%v", err)
				fmt.Print("> ")
				continue
			}

		case "accept":
			if err := cmdAccept(); err != nil {
				printWarning("%v", err)
			}
			fmt.Print("> ")

		case "reject":
			if err := cmdReject(); err != nil {
				printWarning("%v", err)
			}
			fmt.Print("> ")

        case "list":
            printInfo("Seu inventário:")
            for sticker, qty := range cmdListInventory() {
                printInfo(" %s: %d", sticker, qty)
            }
			fmt.Print("> ")

		case "peers":
			node.mu.RLock()
			printInfo("Seus vizinhos:")
			for id, pc := range node.Neighbors {
				printInfo(" %s @ %s", id, pc.Addr)
			}
			fmt.Print("> ")
			node.mu.RUnlock()

        default:
            printWarning("Comando desconhecido: %s", parts[0])
			fmt.Print("> ")
        }
    }
}

func cmdSearch(sticker string) error {
    if !strings.HasPrefix(sticker, "FIG-") {
        return fmt.Errorf("Formato inválido. Use FIG-XX (ex: FIG-23)")
    }
    wantSticker = sticker
    go searchWithRetry()
    return nil
}

func cmdOffer(sticker string) error {
    if !strings.HasPrefix(sticker, "FIG-") {
        return fmt.Errorf("Formato inválido. Use FIG-XX (ex: FIG-23)")
    }
    offerSticker = sticker
    startTradeOffer()
    return nil
}

func cmdAccept() error {
    select {
    case tradeDecision <- "accept":
        return nil
    default:
        return fmt.Errorf("Nenhuma proposta de troca pendente")
    }
}

func cmdReject() error {
    select {
    case tradeDecision <- "reject":
        return nil
    default:
        return fmt.Errorf("Nenhuma proposta de troca pendente")
    }
}

func cmdListInventory() map[string]int {
    node.mu.RLock()
    defer node.mu.RUnlock()
    out := make(map[string]int, len(inventory.Stickers))
    for sticker, qty := range inventory.Stickers {
        out[sticker] = qty
    }
    return out
}

func cmdListPeers() []string {
    node.mu.RLock()
    defer node.mu.RUnlock()
    out := make([]string, 0, len(node.Neighbors))
    for peerID := range node.Neighbors {
        out = append(out, peerID)
    }
    return out
}