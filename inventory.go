package main

import (
    "encoding/json"
    "fmt"
    "os"
)

func initializeInventory(figNumber int) {
    file, err := os.Open(inventoryFile)
    if err != nil {
        // arquivo não existe: inicializa com a figurinha do aluno
        fmt.Println("inventory.json não encontrado. Criando inventário inicial.")
        inventory = Album{
            Stickers: map[string]int{
                fmt.Sprintf("FIG-%02d", figNumber): 28,
            },
        }
        saveInventoryFile()
        return
    }
    defer file.Close()

    if err := json.NewDecoder(file).Decode(&inventory); err != nil {
        fmt.Println("Erro ao ler inventory.json:", err)
        os.Exit(1)
    }

    fmt.Println("Inventário carregado:", inventory.Stickers)
}

func updateInventory(receivedSticker, sentSticker string) {
    inventory.Stickers[receivedSticker]++
    inventory.Stickers[sentSticker]--
    saveInventoryFile()
}

func saveInventoryFile() {
    file, err := os.Create(inventoryFile) // Create trunca e recria o arquivo
    if err != nil {
        fmt.Println("Erro ao abrir inventory.json para escrita:", err)
        return
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(&inventory); err != nil {
        fmt.Println("Erro ao salvar inventory.json:", err)
    }
}