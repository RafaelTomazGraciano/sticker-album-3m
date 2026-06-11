package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// TODO: após o usuário fazer uma escolha, tipo search, offer, etc. espera 5 segundos, caso não receba a resposta envia de novo e espera 10 segundos
// caso nao receba ainda, entao envia mais uma vez e espera 15 segundos, e caso nao receba nenhuma mensagem entao printa que nao recebeu uma respota e entao o usuario pode fazer uma nova acao
func main(){
	// Flags
	idPtr := flag.Int("id", 0, "ID númerico do aluno") 
	peerIP := flag.String("peer", "", "IP do colega para entrar na rede. ex: -peer 192.168.1.10")
	flag.Parse()

	if *idPtr <= 0 {
		fmt.Println("Erro: Você precisa informar um ID de aluno válido e maior que zero.")
		fmt.Println("Use o parâmetro -id. Exemplo: go run . -id 5")
		os.Exit(1)
	}

	peerID := fmt.Sprintf("ALUNO-%02d", *idPtr)
	stickerID := fmt.Sprintf("FIG-%02d", *idPtr)
	
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
		addr := fmt.Sprintf("ws://%s/ws", *peerIP) //Para teste, REMOVER
		//addr := fmt.Sprintf("ws://%s:8080/ws", *peerIP)
		go connectToPeer(addr)
	}

	initializeInventory(*idPtr)

	// Web Socket
	http.HandleFunc("/ws", wsHandler)
	port := findAvailablePort()
	//fmt.Printf("Nó %s iniciado na porta 8080\n", peerID)
	fmt.Printf("Nó %s iniciado na porta %d\n", peerID, port)

	go inputLoop()	

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
	}
}

// apenas para testes. Remover no futuro
func findAvailablePort() int {
    for port := 8080; port <= 8090; port++ {
        ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
        if err == nil {
            ln.Close()
            return port
        }
    }
    fmt.Println("Nenhuma porta disponível entre 8080 e 8090")
    os.Exit(1)
    return 0
}

func inputLoop() {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("Comandos disponíveis:")
    fmt.Println("  search <FIG-XX>   -> buscar uma figurinha")
	fmt.Println("  offer <FIG-XX>    -> oferece uma figurinha sua para troca")
	fmt.Println("  accept            -> aceita a troca entre figurinhas")
	fmt.Println("  reject            -> rejeita a troca entre figurinhas")
    fmt.Println("  list              -> ver seu inventário")

    for {
        fmt.Print("> ")
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
                fmt.Println("Uso: search <FIG-XX>")
                continue
            } else {
				wantSticker = parts[1]
            	startSearch()
			}
		
		case "offer":
			if len(parts) < 2 {
				fmt.Println("Uso: offer <FIG-XX>")
				continue
			}
			offerSticker = parts[1]
			startTradeOffer()
		
		case "accept":
			select {
			case tradeDecision <- "accept":
			default:
				fmt.Println("Nenhuma proposta de troca pendente.")
			}

		case "reject":
			select {
			case tradeDecision <- "reject":
			default:
				fmt.Println("Nenhuma proposta de troca pendente.")
			}

        case "list":
            node.mu.RLock()
            fmt.Println("Seu inventário:")
            for sticker, qty := range node.Inventory {
                fmt.Printf("  %s: %d\n", sticker, qty)
            }
            node.mu.RUnlock()

        default:
            fmt.Printf("Comando desconhecido: %s\n", parts[0])
        }
    }
}