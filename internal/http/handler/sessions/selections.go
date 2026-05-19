package sessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
)

type photographerSelectionResponse struct {
	Session        photographerSelectionSessionResponse  `json:"Session"`
	Payment        *photographerSelectionPaymentResponse `json:"Payment,omitempty"`
	SelectedCount  int32                                 `json:"SelectedCount"`
	SelectedPhotos []photographerSelectedPhotoResponse   `json:"SelectedPhotos"`
}

type photographerSelectionSessionResponse struct {
	ID     uuid.UUID `json:"ID"`
	Status string    `json:"Status"`
}

type photographerSelectionPaymentResponse struct {
	Status      string `json:"Status"`
	AmountCents int32  `json:"AmountCents"`
	PaidAt      any    `json:"PaidAt,omitempty"`
}

type photographerSelectedPhotoResponse struct {
	PhotoID          uuid.UUID  `json:"PhotoID"`
	OriginalFilename string     `json:"OriginalFilename"`
	ThumbURL         string     `json:"ThumbURL,omitempty"`
	Note             string     `json:"Note"`
	SelectedAt       any        `json:"SelectedAt"`
	FinalID          *uuid.UUID `json:"FinalID,omitempty"`
	FinalUploaded    bool       `json:"FinalUploaded"`
}

const presignedGetTTL = 30 * time.Minute

func (h *Handler) GetSessionSelections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID, err := uuid.Parse(r.PathValue("sessionId"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	if err := guard.EnsureSessionOwner(ctx, h.sessionsRepo, sessionID, userID); err != nil {
		if errors.Is(err, guard.ErrSessionNotAccessible) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	overview, err := h.sessionsRepo.GetPhotographerSelectionOverview(ctx, sessionID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	photos := make([]photographerSelectedPhotoResponse, 0, len(overview.Photos))

	for _, photo := range overview.Photos {
		thumbURL := ""

		if photo.ThumbKey != "" {
			signedURL, err := h.storage.PresignedGetObject(ctx, photo.ThumbKey, presignedGetTTL)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			thumbURL = signedURL
		}

		photos = append(photos, photographerSelectedPhotoResponse{
			PhotoID:          photo.PhotoID,
			OriginalFilename: photo.OriginalFilename,
			ThumbURL:         thumbURL,
			Note:             photo.Note,
			SelectedAt:       photo.SelectedAt,
			FinalID:          photo.FinalID,
			FinalUploaded:    photo.FinalUploaded,
		})
	}

	var payment *photographerSelectionPaymentResponse

	if overview.PaymentStatus != "" {
		payment = &photographerSelectionPaymentResponse{
			Status:      overview.PaymentStatus,
			AmountCents: overview.PaymentAmountCents,
			PaidAt:      overview.PaymentPaidAt,
		}
	}

	response := photographerSelectionResponse{
		Session: photographerSelectionSessionResponse{
			ID:     overview.SessionID,
			Status: overview.SessionStatus,
		},
		Payment:        payment,
		SelectedCount:  overview.SelectedCount,
		SelectedPhotos: photos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
