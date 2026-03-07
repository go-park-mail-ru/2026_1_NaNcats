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

func (h *imageHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер аватара в 4Мб
	err := r.ParseMultipartForm(4 << 20)
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

	savedPath, err := h.imageUC.UploadImage(ctx, fileBytes, header.Filename, "users/avatars")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]string{
		"url": "/api/images/" + savedPath,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *imageHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер логотипа в 6Мб
	err := r.ParseMultipartForm(6 << 20)
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

	savedPath, err := h.imageUC.UploadImage(ctx, fileBytes, header.Filename, "restaurants/logos")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]string{
		"url": "/api/images/" + savedPath,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *imageHandler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер баннера в 10Мб
	err := r.ParseMultipartForm(10 << 20)
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

	savedPath, err := h.imageUC.UploadImage(ctx, fileBytes, header.Filename, "restaurants/banners")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]string{
		"url": "/api/images/" + savedPath,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *imageHandler) UploadDishes(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер логотипа в 6Мб
	err := r.ParseMultipartForm(6 << 20)
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

	savedPath, err := h.imageUC.UploadImage(ctx, fileBytes, header.Filename, "dishes")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := map[string]string{
		"url": "/api/images/" + savedPath,
	}

	response.JSON(w, http.StatusOK, resp)
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
