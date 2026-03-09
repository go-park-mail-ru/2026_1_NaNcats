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

// Download godoc
// @Summary 		Получение фото по URL
// @Description		Возвращает фотографию, которая храниться по указанному в GET-запросе URL
// @Tags			images
// @Produce				image/png
// @Produce				image/jpg
// @Produce				image/jpeg
// @Produce				image/webp
// @Param				filepath string true "Относительный путь до файла (например: restaurants/logos/123.png)"
// @Success				200		{file}  file				"Сырые байты изображения"
// @Failure				404		{object}  response.ErrorResponse	"Изображение не найдено"
// @Failure				500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/images/{filepath} [get]
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
