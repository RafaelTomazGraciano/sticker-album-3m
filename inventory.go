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
        printSystem("inventory.json não encontrado. Criando inventário inicial")
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
        printError("Erro ao ler inventory.json: %v", err)
        os.Exit(1)
    }

    printSystem("Inventário carregado: %v", inventory.Stickers)
}

func updateInventory(receivedSticker, sentSticker string) {
    node.mu.Lock()
    defer node.mu.Unlock()

    if inventory.Stickers[sentSticker] <= 0 {
        printError("Tentativa de enviar %s sem possuir o item (quantidade: %d)", sentSticker, inventory.Stickers[sentSticker])
        return
    }

    if _, ok := inventory.Stickers[receivedSticker]; !ok {
        inventory.Stickers[receivedSticker] = 0
    }

    inventory.Stickers[receivedSticker]++
    inventory.Stickers[sentSticker]--
    saveInventoryFile()
}

func saveInventoryFile() {
    file, err := os.Create(inventoryFile) // Create trunca e recria o arquivo
    if err != nil {
        printError("Erro ao abrir inventory.json para escrita: %v", err)
        return
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(&inventory); err != nil {
        printError("Erro ao salvar inventory.json: %v", err)
    }
}