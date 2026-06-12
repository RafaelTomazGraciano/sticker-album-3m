# Sticker Album P2P
 
Sistema de troca de figurinhas em rede P2P não estruturada, desenvolvido em Go. Cada nó representa um aluno, mantém seu próprio inventário e se comunica com outros nós via WebSocket usando mensagens JSON.

## Pré-requisitos

* [Go](https://go.dev/doc/install) versão 1.26.4 ou superior

## Como executar o projeto

Sincronize as dependências do projeto:

```bash
go mod tidy
```

Para executar o projeto, utilize o comando abaixo, substituindo `ID` pelo seu número na lista de chamada e `IP` pelo endereço IP de um nó já existente na rede:

```bash
   go run . -id ID -peer IP
```

## Comandos disponíveis
 
| Comando | Descrição |
|---|---|
| `search <FIG-XX>` | Busca uma figurinha na rede por inundação |
| `offer <FIG-XX>` | Propõe uma troca após localizar a figurinha |
| `accept` | Aceita a proposta de troca recebida |
| `reject` | Recusa a proposta de troca recebida |
| `list` | Exibe o inventário local |
 
## Fluxo de troca
 
1. `search FIG-XX` — localiza quem possui a figurinha
2. `offer FIG-YY` — propõe a troca oferecendo uma figurinha sua
3. O outro nó recebe a proposta e digita `accept` ou `reject`
4. Se aceita, ambos os inventários são atualizados automaticamente

## Teste local

Para executar os testes, abra um terminal e execute o comando abaixo:

```bash
go run . -id 23
```

Abra outro terminal e execute o comando abaixo:

```bash
go run . -id 5 -peer 127.0.0.1:PORTA
```

Faça uma busca pela figurinha 23:

```bash
search FIG-23
```