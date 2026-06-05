package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main(){
	// Flag id
	idPtr := flag.Int("id", 0, "ID númerico do aluno") 
	flag.Parse()

	if *idPtr <= 0{
		fmt.Println("Erro: Você precisa informar um ID de aluno válido e maior que zero.")
		fmt.Println("Use o parâmetro -id. Exemplo: go run . -id 5")
		os.Exit(1)
	}

	idAluno := *idPtr
	peerID := fmt.Sprintf("ALUNO-%02d", idAluno)
	stickerID := fmt.Sprintf("FIG-%02d", idAluno)
	fmt.Printf("Peer: %s, Sticker: %s\n", peerID, stickerID)

	// Web Socket
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("WebSocket server started on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}