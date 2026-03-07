package handler

import (
	"io"
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

func (h *imageHandler) imageUpload(w http.ResponseWriter, r *http.Request, maxMemory int64, path string) {
	// Ограничиваем размер
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "File is too big")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Not found file in 'photo'")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "File reading error")
		return
	}

	ctx := r.Context()

	savedPath, err := h.imageUC.UploadImage(ctx, fileBytes, header.Filename, path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]string{
		"url": "/api/images/" + savedPath,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *imageHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	h.imageUpload(w, r, 4<<20, "users/avatars")
}

func (h *imageHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	h.imageUpload(w, r, 6<<20, "restaurants/logos")
}

func (h *imageHandler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	h.imageUpload(w, r, 10<<20, "restaurants/banners")
}

func (h *imageHandler) UploadDishes(w http.ResponseWriter, r *http.Request) {
	h.imageUpload(w, r, 6<<20, "dishes")
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
