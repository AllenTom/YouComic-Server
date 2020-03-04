package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allentom/youcomic-api/auth"
	ApiError "github.com/allentom/youcomic-api/error"
	ApplicationError "github.com/allentom/youcomic-api/error"
	"github.com/allentom/youcomic-api/model"
	"github.com/allentom/youcomic-api/permission"
	"github.com/allentom/youcomic-api/serializer"
	"github.com/allentom/youcomic-api/services"
	"github.com/allentom/youcomic-api/utils"
	"github.com/allentom/youcomic-api/validate"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

type CreateBookRequestBody struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
}

// create book handler
//
// path: /books
//
// method: post
var CreateBookHandler gin.HandlerFunc = func(context *gin.Context) {
	var requestBody CreateBookRequestBody
	DecodeJsonBody(context, &requestBody)

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApplicationError.UserAuthFailError, nil)
		return
	}

	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.CreateBookPermissionName, UserId: claims.UserId},
	)

	validate.RunValidatorsAndRaiseApiError(context,
		&validate.UniqBookNameValidator{Value: requestBody.Name},
	)

	err, book := services.CreateBook(requestBody.Name)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//serializer response
	template := serializer.BaseBookTemplate{}
	RenderTemplate(context, &template, *book)
	context.JSON(http.StatusCreated, template)
}

type UpdateBookRequestBody struct {
	Id   int
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
}

// update book handler
//
// path: /book/:id
//
// method: patch
var UpdateBookHandler gin.HandlerFunc = func(context *gin.Context) {

	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	//check permission
	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.UserId},
	)

	requestBody := UpdateBookRequestBody{}
	DecodeJsonBody(context, &requestBody)

	//validate
	validate.RunValidatorsAndRaiseApiError(context,
		&validate.StringLengthValidator{Value: requestBody.Name, LessThan: 256, GreaterThan: 0, FieldName: "BookName"},
	)

	book := &model.Book{}
	err = AssignUpdateModel(&requestBody, book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	book.ID = uint(id)

	err = services.UpdateBook(book, "Name")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.GetBook(book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	template := &serializer.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSON(http.StatusOK, template)
}

// get book list handler
//
// path: /books
//
// method: get
var BookListHandler gin.HandlerFunc = func(context *gin.Context) {
	//get page
	pagination := DefaultPagination{}
	pagination.Read(context)
	//get filter
	var books []model.Book
	queryBuilder := services.BooksQueryBuilder{}
	queryBuilder.SetPageFilter(pagination.Page, pagination.PageSize)

	filterMapping := []FilterMapping{
		{
			Lookup: "id",
			Method: "InId",
			Many:   true,
		},
		{
			Lookup: "name",
			Method: "SetNameFilter",
			Many:   true,
		},
		{
			Lookup: "order",
			Method: "SetOrderFilter",
			Many:   false,
		},
		{
			Lookup: "collection",
			Method: "SetCollectionQueryFilter",
			Many:   true,
		},
		{
			Lookup: "tag",
			Method: "SetTagQueryFilter",
			Many:   true,
		},
		{
			Lookup: "startTime",
			Method: "SetStartTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "endTime",
			Method: "SetEndTimeQueryFilter",
			Many:   false,
		},
		{
			Lookup: "nameSearch",
			Method: "SetNameSearchQueryFilter",
			Many:   false,
		},
	}
	for _, filter := range filterMapping {
		utils.FilterByParam(context, filter.Lookup, &queryBuilder, filter.Method, filter.Many)
	}

	count, err := queryBuilder.ReadModels(&books)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	with := context.GetStringSlice("with")
	result := serializer.SerializeMultipleTemplate(books, &serializer.BaseBookTemplate{}, map[string]interface{}{"with": with})
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     pagination.Page,
		"pageSize": pagination.PageSize,
		"count":    count,
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}

// delete book handler
//
// path: /book/:id
//
// method: delete
var DeleteBookHandler gin.HandlerFunc = func(context *gin.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.RequestPathError, nil)
		return
	}

	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	//check permission
	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.DeleteBookPermissionName, UserId: claims.UserId},
	)

	book := &model.Book{}
	book.ID = uint(id)
	err = services.DeleteById(&book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type BatchRequestBody struct {
	Create []*CreateBookRequestBody `json:"create"`
	Update []*UpdateBookRequestBody `json:"update"`
	Delete []int                    `json:"delete"`
}

// books action handler
//
// path: /books/batch
//
// method: post
var BookBatchHandler gin.HandlerFunc = func(context *gin.Context) {
	requestBody := BatchRequestBody{}
	DecodeJsonBody(context, &requestBody)

	//create action
	claims, err := auth.ParseAuthHeader(context)
	if err != nil {
		ApiError.RaiseApiError(context, ApiError.UserAuthFailError, nil)
		return
	}

	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.CreateBookPermissionName, UserId: claims.UserId},
	)
	booksToCreate := make([]model.Book, 0)
	for _, requestBook := range requestBody.Create {
		book := model.Book{}
		err = copier.Copy(&book, &requestBook)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		booksToCreate = append(booksToCreate, book)
	}
	err = services.CreateBooks(booksToCreate)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//update
	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.UpdateBookPermissionName, UserId: claims.UserId},
	)
	booksToUpdate := make([]model.Book, 0)
	for _, updateBook := range requestBody.Update {
		book := model.Book{}
		err = AssignUpdateModel(&updateBook, &book)
		book.ID = uint(updateBook.Id)
		if err != nil {
			ApiError.RaiseApiError(context, err, nil)
			return
		}
		booksToUpdate = append(booksToUpdate, book)
	}
	err = services.UpdateBooks(booksToUpdate, "Name")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//delete
	permission.ChePermissionAndServerError(context,
		&permission.StandardPermissionChecker{PermissionName: permission.DeleteBookPermissionName, UserId: claims.UserId},
	)
	err = services.DeleteBooks(requestBody.Delete...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type AddTagToBookRequestBody struct {
	Tags []int `json:"tags"`
}

var BookTagBatch gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	requestBody := AddTagToBookRequestBody{}
	err = context.ShouldBindJSON(&requestBody)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.AddTagToBook(id, requestBody.Tags...)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

func SaveCover(context *gin.Context, bookId int, file *multipart.FileHeader) (error, string, string) {
	err, storePath := services.GetBookPath(bookId)
	if err != nil {
		return err, "", ""
	}
	fileExt := filepath.Ext(file.Filename)
	coverImageFilePath := filepath.Join(storePath, fmt.Sprintf("cover%s", fileExt))
	err = context.SaveUploadedFile(file, coverImageFilePath)
	if err != nil {
		return err, "", ""
	}

	//generate thumbnail image
	thumbnailImageFile, err := os.Open(coverImageFilePath)
	if err != nil {
		return err, "", ""
	}
	var thumbnailImage image.Image
	if fileExt == ".png" {
		thumbnailImage, err = png.Decode(thumbnailImageFile)
	}
	if fileExt == ".jpg" {
		thumbnailImage, err = jpeg.Decode(thumbnailImageFile)
	}
	if err != nil {
		return err, "", ""
	}
	resizeImage := resize.Thumbnail(480, 480, thumbnailImage, resize.Lanczos3)
	thumbnailImagePath := filepath.Join(storePath, fmt.Sprintf("cover_thumbnail%s", fileExt))
	output, err := os.Create(thumbnailImagePath)
	if err != nil {
		return err, "", ""
	}

	defer thumbnailImageFile.Close()
	defer output.Close()

	err = jpeg.Encode(output, resizeImage, nil)
	if err != nil {
		return err, "", ""
	}
	return nil, coverImageFilePath, thumbnailImagePath

}

var AddBookCover gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	form, err := context.MultipartForm()
	if form == nil {
		ApiError.RaiseApiError(context, errors.New("form not found"), nil)
		return
	}
	if _, isFileExistInForm := form.File["image"]; !isFileExistInForm {
		ApiError.RaiseApiError(context, errors.New("no such file in form"), nil)
		return
	}
	//update database
	book := model.Book{Model: gorm.Model{ID: uint(id)}}
	err = services.GetBook(&book)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	fileHeader := form.File["image"][0]
	err, coverImageFilePath, _ := SaveCover(context, id, fileHeader)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	book.Cover = filepath.Base(coverImageFilePath)
	err = services.UpdateModel(&book, "Cover")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	template := &serializer.BaseBookTemplate{}
	RenderTemplate(context, template, book)
	context.JSON(http.StatusOK, template)
}

var AddBookPages gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	form, err := context.MultipartForm()
	if form == nil {
		context.JSON(http.StatusOK, "template")
		return
	}
	re, err := regexp.Compile(`^page_(\d+)$`)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err, storePath := services.GetBookPath(id)
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	createPages := make([]model.Page, 0)
	for fileField, file := range form.File {
		if re.MatchString(fileField) {
			matchGroups := re.FindAllStringSubmatch(fileField, 1)
			if len(matchGroups) > 0 && len(matchGroups[0]) > 1 {
				orderStr := matchGroups[0][1]
				order, err := strconv.Atoi(orderStr)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				//store
				storeFileHeader := file[0]
				fileExt := path.Ext(storeFileHeader.Filename)
				storeFileName := fmt.Sprintf("page_%d%s", order, fileExt)
				err = context.SaveUploadedFile(storeFileHeader, fmt.Sprintf("%s/%s", storePath, storeFileName))
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				page := &model.Page{Path: storeFileName, Order: order, BookId: id}
				err = services.CreateModel(page)
				if err != nil {
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				createPages = append(createPages, *page)
			}
		}
	}

	result := serializer.SerializeMultipleTemplate(createPages, &serializer.BasePageTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     1,
		"pageSize": len(createPages),
		"count":    len(createPages),
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}

var GetBookTags gin.HandlerFunc = func(context *gin.Context) {
	var err error
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tags, err := services.GetBookTag(uint(id))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	result := serializer.SerializeMultipleTemplate(tags, &serializer.BaseTagTemplate{}, nil)
	responseBody := serializer.DefaultListContainer{}
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     1,
		"pageSize": len(tags),
		"count":    len(tags),
		"url":      context.Request.URL,
	})
	context.JSON(http.StatusOK, responseBody)
}

var DeleteBookTag gin.HandlerFunc = func(context *gin.Context) {
	id, err := GetLookUpId(context, "id")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tagId, err := GetLookUpId(context, "tag")
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	err = services.RemoveTagFromBook(uint(id), uint(tagId))
	if err != nil {
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	ServerSuccessResponse(context)
}

type UploadBookRequestBody struct {
	Name  string `form:"name"`
	Tags  string `form:"tags"`
	Pages string `form:"pages"`
	Cover string `form:"cover"`
}

var CreateBook gin.HandlerFunc = func(context *gin.Context) {
	var requestBody UploadBookRequestBody
	err := context.ShouldBind(&requestBody)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err, book := services.CreateBook(requestBody.Name)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}
	tagToAdd := make([]*model.Tag, 0)
	err = json.Unmarshal([]byte(requestBody.Tags), &tagToAdd)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	err = services.AddOrCreateTagToBook(book, tagToAdd)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	//handle with pages
	form, _ := context.MultipartForm()
	files := form.File["image"]
	pageFilenames := make([]string, 0)
	err = json.Unmarshal([]byte(requestBody.Pages), &pageFilenames)
	if err != nil {
		logrus.Error(err)
		ApiError.RaiseApiError(context, err, nil)
		return
	}

	for _, pageFilename := range pageFilenames {
		for pageIdx, file := range files {
			if pageFilename == file.Filename {
				storePath, err := SavePageFile(context, file, int(book.ID), pageIdx)
				if err != nil {
					logrus.Error(err)
					ApiError.RaiseApiError(context, err, nil)
					return
				}
				err = services.CreatePage(&model.Page{Order: pageIdx, Path: filepath.Base(storePath), BookId: int(book.ID)})
				if err != nil {
					logrus.Error(err)
					ApiError.RaiseApiError(context, err, nil)
					return
				}
			}
		}
	}

	for _, file := range files {
		if file.Filename == requestBody.Cover {
			//save cover
			err, coverPath, _ := SaveCover(context, int(book.ID), file)
			if err != nil {
				logrus.Error(err)
				ApiError.RaiseApiError(context, err, nil)
				return
			}
			book.Cover = filepath.Base(coverPath)
			err = services.UpdateBook(book, "Cover")
			if err != nil {
				logrus.Error(err)
				ApiError.RaiseApiError(context, err, nil)
				return
			}
		}
	}

	template := &serializer.BaseBookTemplate{}
	RenderTemplate(context, template, *book)
	context.JSON(http.StatusOK, template)
}