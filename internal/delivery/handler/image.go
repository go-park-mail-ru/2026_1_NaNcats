package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

type imageHandler struct {
	imageUC usecase.ImageUseCase
}

// функция-конструтор хендлера
func NewImageHandler(iuc usecase.ImageUseCase) *imageHandler {
	return &imageHandler{
		imageUC: iuc,
	}
}

func (h *imageHandler) Download(w http.ResponseWriter, r *http.Request) {
	filePath := r.PathValue("filepath")

	ctx := r.Context()

	imageBytes, err := h.imageUC.GetImage(ctx, filePath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Server can not read file")
		return
	}

	contentType := http.DetectContentType(imageBytes)
	w.Header().Set("Content-Type", contentType)
	w.Write(imageBytes)
}
