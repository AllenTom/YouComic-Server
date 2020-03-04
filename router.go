package main

import (
	"github.com/allentom/youcomic-api/controller"
	"github.com/gin-gonic/gin"
)

func SetRouter(engine *gin.Engine) {
	engine.POST("/books", controller.CreateBookHandler)
	engine.POST("/books/upload", controller.CreateBook)
	engine.PATCH("/book/:id", controller.UpdateBookHandler)
	engine.PUT("/book/:id/tags", controller.BookTagBatch)
	engine.PUT("/book/:id/cover", controller.AddBookCover)
	engine.PUT("/book/:id/pages", controller.AddBookPages)
	engine.DELETE("/book/:id/tag/:tag", controller.DeleteBookTag)
	engine.GET("/books", controller.BookListHandler)
	engine.DELETE("/book/:id", controller.DeleteBookHandler)
	engine.POST("/books/batch", controller.BookBatchHandler)
	engine.POST("/pages", controller.PageUploadHandler)
	engine.PATCH("/page/:id", controller.UpdatePageHandler)
	engine.DELETE("/page/:id", controller.DeletePageHandler)
	engine.GET("/pages", controller.PageListHandler)
	engine.POST("/pages/batch", controller.BatchPageHandler)
	engine.POST("/tags", controller.CreateTagHandler)
	engine.POST("/tags/batch", controller.BatchTagHandler)
	engine.GET("/tags", controller.TagListHandler)
	engine.GET("/book/:id/tags", controller.GetBookTags)
	engine.GET("/tag/:id/books", controller.TagBooksHandler)
	engine.PUT("/tag/:id/books", controller.AddBooksToTagHandler)
	engine.DELETE("/tag/:id/books", controller.RemoveBooksFromTagHandler)
	engine.POST("/user/register", controller.RegisterUserHandler)
	engine.POST("/user/auth", controller.LoginUserHandler)
	engine.GET("/user/:id", controller.GetUserHandler)
	engine.POST("/collections", controller.CreateCollectionHandler)
	engine.GET("/collections", controller.CollectionsListHandler)
	engine.PUT("/collection/:id/books", controller.AddToCollectionHandler)
	engine.DELETE("/collection/:id/books", controller.DeleteFromCollectionHandler)
	engine.PUT("/collection/:id/users", controller.AddUsersToCollectionHandler)
	engine.DELETE("/collection/:id/users", controller.DeleteUsersFromCollectionHandler)
	engine.DELETE("/collection/:id", controller.DeleteCollectionHandler)
}