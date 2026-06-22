package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const totalStickers = 28

func registerAPIRoutes() {
	http.HandleFunc("/api/status", handleAPIStatus)
	http.HandleFunc("/api/peers", handleAPIPeers)
	http.HandleFunc("/api/album", handleAPIAlbum)
	http.HandleFunc("/api/search", handleAPISearch)
	http.HandleFunc("/api/offer", handleAPIOffer)
	http.HandleFunc("/api/pending-offer", handleAPIPendingOffer)
	http.HandleFunc("/api/trade/decision", handleAPITradeDecision)
	http.HandleFunc("/api/logs", handleAPILogs)
	http.Handle("/", http.FileServer(http.Dir("web")))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":        node.ID,
		"ip":        localIP,
		"inventory": cmdListInventory(),
	})
}

func handleAPIPeers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"peers": cmdListPeers(),
	})
}

type albumSticker struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

func handleAPIAlbum(w http.ResponseWriter, r *http.Request) {
	inv := cmdListInventory()
	stickers := make([]albumSticker, 0, totalStickers)
	for i := 1; i <= totalStickers; i++ {
		id := fmt.Sprintf("FIG-%02d", i)
		stickers = append(stickers, albumSticker{ID: id, Qty: inv[id]})
	}
	writeJSON(w, http.StatusOK, map[string]any{"stickers": stickers})
}

type stickerRequest struct {
	Sticker string `json:"sticker"`
}

func handleAPISearch(w http.ResponseWriter, r *http.Request) {
	var req stickerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "json inválido")
		return
	}
	if err := cmdSearch(req.Sticker); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func handleAPIOffer(w http.ResponseWriter, r *http.Request) {
	var req stickerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "json inválido")
		return
	}
	if err := cmdOffer(req.Sticker); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func handleAPIPendingOffer(w http.ResponseWriter, r *http.Request) {
	offer := getPendingOffer()
	if offer == nil {
		writeJSON(w, http.StatusOK, map[string]bool{"pending": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pending": true,
		"from":    offer.OriginPeerID,
		"offers":  offer.OfferSticker,
		"wants":   offer.WantSticker,
	})
}

type tradeDecisionRequest struct {
	Decision string `json:"decision"`
}

func handleAPITradeDecision(w http.ResponseWriter, r *http.Request) {
	var req tradeDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "json inválido")
		return
	}

	var err error
	switch req.Decision {
	case "accept":
		err = cmdAccept()
	case "reject":
		err = cmdReject()
	default:
		writeAPIError(w, http.StatusBadRequest, "decision deve ser 'accept' ou 'reject'")
		return
	}

	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func handleAPILogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"logs": getLogs()})
}
