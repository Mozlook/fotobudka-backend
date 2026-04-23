package jobsworker

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mozlook/fotobudka-backend/internal/deliveries"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	"github.com/google/uuid"
)

func sanitizeZIPEntryName(name string, fallbackID uuid.UUID) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." || name == "/" {
		return fallbackID.String()
	}
	return name
}

func deliveryZIPObjectKey(sessionID uuid.UUID, version int32) string {
	return fmt.Sprintf("sessions/%s/deliveries/v%d.zip", sessionID, version)
}

func (w *Worker) handleGenerateDeliveryZIP(ctx context.Context, job jobs.Job) error {
	var payload deliveries.GenerateDeliveryZIPPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal generate_delivery_zip payload: %w", err)
	}

	if payload.SessionID == uuid.Nil {
		return fmt.Errorf("session_id cannot be empty")
	}
	if payload.DeliveryID == uuid.Nil {
		return fmt.Errorf("delivery_id cannot be empty")
	}

	delivery, err := w.deliveriesRepo.GetDeliveryByID(ctx, payload.DeliveryID)
	if err != nil {
		return retryable(fmt.Errorf("get delivery by id: %w", err))
	}

	if delivery.SessionID != payload.SessionID {
		return fmt.Errorf("delivery does not belong to session")
	}
	if delivery.Status != "generating" {
		return fmt.Errorf("delivery is not in generating status")
	}

	finals, err := w.finalPhotosRepo.ListFinalPhotosForDelivery(ctx, payload.SessionID)
	if err != nil {
		return retryable(fmt.Errorf("list final photos for delivery: %w", err))
	}
	if len(finals) == 0 {
		return fmt.Errorf("no final photos for delivery")
	}

	tmpFile, err := os.CreateTemp("", "fotobudka-delivery-*.zip")
	if err != nil {
		return retryable(fmt.Errorf("create temp zip file: %w", err))
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	zw := zip.NewWriter(tmpFile)

	usedNames := make(map[string]int, len(finals))

	for _, final := range finals {
		obj, err := w.storage.GetObject(ctx, final.FinalKey)
		if err != nil {
			if errors.Is(err, storage.ErrObjectNotFound) {
				return fmt.Errorf("final object not found for photo %s", final.PhotoID)
			}
			return retryable(fmt.Errorf("get final object for photo %s: %w", final.PhotoID, err))
		}

		entryName := sanitizeZIPEntryName(final.OriginalFilename, final.PhotoID)
		if count, exists := usedNames[entryName]; exists {
			count++
			usedNames[entryName] = count

			ext := filepath.Ext(entryName)
			base := strings.TrimSuffix(entryName, ext)
			entryName = fmt.Sprintf("%s_%d%s", base, count, ext)
		} else {
			usedNames[entryName] = 1
		}

		writer, err := zw.Create(entryName)
		if err != nil {
			_ = obj.Close()
			return retryable(fmt.Errorf("create zip entry for photo %s: %w", final.PhotoID, err))
		}

		if _, err := io.Copy(writer, obj); err != nil {
			_ = obj.Close()
			return retryable(fmt.Errorf("copy final object to zip for photo %s: %w", final.PhotoID, err))
		}

		_ = obj.Close()
	}

	if err := zw.Close(); err != nil {
		return retryable(fmt.Errorf("close zip writer: %w", err))
	}

	info, err := tmpFile.Stat()
	if err != nil {
		return retryable(fmt.Errorf("stat zip file: %w", err))
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		return retryable(fmt.Errorf("rewind zip file: %w", err))
	}

	zipKey := deliveryZIPObjectKey(payload.SessionID, delivery.Version)

	if err := w.storage.PutObject(ctx, zipKey, "application/zip", tmpFile, info.Size()); err != nil {
		return retryable(fmt.Errorf("put zip object: %w", err))
	}

	if err := w.deliveriesRepo.MarkDeliveryReadyAndSessionDelivered(ctx, payload.DeliveryID, payload.SessionID, zipKey, info.Size()); err != nil {
		return retryable(fmt.Errorf("mark delivery ready and session delivered: %w", err))
	}

	return nil
}
