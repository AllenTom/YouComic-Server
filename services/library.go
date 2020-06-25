package services

import (
	"encoding/json"
	"fmt"
	appconfig "github.com/allentom/youcomic-api/config"
	"github.com/allentom/youcomic-api/database"
	"github.com/allentom/youcomic-api/model"
	"io/ioutil"
	"os"
	"path"
)

func GetLibraryById(id uint) (model.Library, error) {
	var library model.Library
	err := database.DB.Find(&library, id).Error
	return library, err
}

func CreateLibrary(name string, path string) (*model.Library, error) {
	// create library with path
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	newLibrary := &model.Library{Name: name, Path: path}
	err = database.DB.Create(newLibrary).Error
	return newLibrary, err
}

type LibraryExportConfig struct {
	Name  string `json:"name"`
	Books []struct {
		Name string `json:"name,omitempty"`
		Path string `json:"path"`
		Tags []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"tags"`
		Pages []struct {
			Path  string `json:"path"`
			Order int    `json:"order"`
		} `json:"pages"`
		Cover string `json:"cover"`
	} `json:"books"`
}

func ImportLibrary(libraryPath string) error {
	file, err := ioutil.ReadFile(path.Join(libraryPath, "library_export.json"))
	if err != nil {
		return err
	}
	config := LibraryExportConfig{}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	// register new library
	library, err := CreateLibrary(config.Name, libraryPath)
	if err != nil {
		return err
	}
	// add library book
	for _, bookConfig := range config.Books {
		book := model.Book{Name: bookConfig.Name, Path: bookConfig.Path, LibraryId: library.ID, Cover: bookConfig.Cover}
		err = database.DB.Create(&book).Error
		if err != nil {
			return err
		}
		//generate cover thumbnail
		coverAbsolutePath := path.Join(libraryPath,bookConfig.Path,bookConfig.Cover)
		coverThumbnailStorePath := path.Join(appconfig.Config.Store.Root,"generate",fmt.Sprintf("%d",book.ID))
		_,err = GenerateCoverThumbnail(coverAbsolutePath,coverThumbnailStorePath)
		if err != nil {
			return err
		}
		for _, tagConfig := range bookConfig.Tags {
			tag := model.Tag{}
			err = database.DB.FirstOrCreate(&tag,model.Tag{Name: tagConfig.Name,Type: tagConfig.Type}).Error
			if err != nil {
				return err
			}
			database.DB.Model(&book).Association("Tags").Append(&tag)
		}
		for _, pageConfig := range bookConfig.Pages {
			page := model.Page{Order: pageConfig.Order, Path: pageConfig.Path, BookId: int(book.ID)}
			err = database.DB.Create(&page).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}
