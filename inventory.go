package main

import(
	"fmt"
	"os"
	"encoding/json"
	"type"
)

const inventoryFile = "inventory.json"
var inventory Album = {};

func initializeInventory(figNumber int){
	file, err := os.Open(inventoryFile)
	if err != nil {
		fmt.Println("Erro ao ler o arquivo inventory.json. Criando um inventário limpo")
		file, e := os.Create(inventoryFile)
		if e != nil {
			fmt.Println("Falha ao criar o inventário: %s", e)
			os.Exit(1)
		}
	} 
	//TODO: ler o json e salvar no inventory

	defer file.Close()
}

func updateInventory(recievedSticker, sentSticker string){
	
}

func saveInventoryFile(){
	bytesWritten, err := file.WriteString(inventory)
	if err != nil {
		fmt.Println("Falha ao escrever no arquivo: %s", err)
		os.Exit(1)
	}
}
