package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func initializeInventory(figNumber int) {
	file, err := os.Open(inventoryFile)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo inventory.json. Criando um inventário limpo")
		_, e := os.Create(inventoryFile)
		if e != nil {
			fmt.Println("Falha ao criar o inventário: %s", e)
			os.Exit(1)
		}
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&inventory)
	if err != nil {
		fmt.Println("Erro ao ler o json")
	}
	fmt.Println(inventory)

	defer file.Close()
}

func updateInventory(recievedSticker, sentSticker string) {
	if _, ok := inventory.Stickers[recievedSticker]; !ok {
		inventory.Stickers[recievedSticker] = 1
	} else {
		inventory.Stickers[recievedSticker]++
	}

	inventory.Stickers[sentSticker]--
	saveInventoryFile()
}

func saveInventoryFile() {
	file, err := os.Open(inventoryFile)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo inventory.json. Razão: ", err)
	}
	decoder := json.NewEncoder(file)
	err = decoder.Encode(&inventory)
	if err != nil {
		fmt.Println("Erro ao codificar o json")
	}
	fmt.Println(inventory)

	defer file.Close()
}
