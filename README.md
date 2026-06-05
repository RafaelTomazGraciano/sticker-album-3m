# Sticker Album 3M

## Pré-requisitos

* [Go](https://go.dev/doc/install) versão 1.26.4

## Como executar o projeto

Sincronize as dependências do projeto:

```bash
go mod tidy
```
Para executar o projeto, utilize o comando abaixo, substituindo `ID` pelo seu número na lista de chamada e `IP` pelo endereço IP de um nó já existente na rede:

```bash
   go run . -id ID -peer IP
```

## Testes

Para executar os testes, abra um terminal e execute o comando abaixo:

```bash
go run . -id 23
```

Abra outro terminal e execute o comando abaixo:

```bash
go run . -id 5 -peer 127.0.0.1
```

Faça uma busca pela figurinha 23:

```bash
search FIG-23
```