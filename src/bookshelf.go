package src

import (
	"fmt"
	"github.com/VeronicaAlexia/BoluobaoAPI/Template"
	sfbook "github.com/VeronicaAlexia/BoluobaoAPI/boluobao/book"
	"github.com/VeronicaAlexia/BoluobaoAPI/boluobao/bookshelf"
	HbookerAPI "github.com/VeronicaAlexia/HbookerAPI/ciweimao/bookshelf"
	"github.com/VeronicaAlexia/pineapple-backups/config"
	"github.com/VeronicaAlexia/pineapple-backups/pkg/tools"
	"strconv"
	"strings"
)

func sfacg_bookshelf() (map[int][]map[string]string, error) {
	response := bookshelf.GET_BOOK_SHELF_INFORMATION()
	bookshelf_info := make(map[int][]map[string]string)
	if response.Status.HTTPCode != 200 {
		return nil, fmt.Errorf(response.Status.Msg.(string))
	}
	for index, value := range response.Data {
		fmt.Println("书架号:", index, "\t\t\t书架名:", value.Name)
		var bookshelf_info_list []map[string]string
		for _, book := range value.Expand.Novels {
			bookshelf_info_list = append(bookshelf_info_list,
				map[string]string{"novel_name": book.NovelName, "novel_id": strconv.Itoa(book.NovelID)},
			)
		}
		bookshelf_info[index] = bookshelf_info_list
	}
	return bookshelf_info, nil
}

type Bookshelf struct {
	ShelfIndex int
}

func (bs *Bookshelf) ShowBookshelf(shelf Template.InfoData) {
	for index, value := range shelf.Data {
		fmt.Println("书架号:", index, "\t\t\t书架名:", value.Name)
	}

}

func (bs *Bookshelf) ChoiceBookshelf(shelf Template.InfoData) {
	for _, book := range shelf.Data[bs.ShelfIndex].Expand.Novels {
		fmt.Println("小说名:", book.NovelName, "\t\t\t小说ID:", book.NovelID)
	}
	choice := tools.InputStr("是继续进行书架下载?[y/a/enter]")
	if strings.ToLower(choice) == "y" {
		BookIndex := tools.InputInt(">", len(shelf.Data[bs.ShelfIndex].Expand.Novels))
		sfbook.GET_BOOK_INFORMATION(strconv.Itoa(shelf.Data[bs.ShelfIndex].Expand.Novels[BookIndex].NovelID))
	} else if strings.ToLower(choice) == "a" {
		for _, book := range shelf.Data[bs.ShelfIndex].Expand.Novels {
			sfbook.GET_BOOK_INFORMATION(strconv.Itoa(book.NovelID))
		}
	} else {
		fmt.Println("已退出书架下载")
	}

}

func (bs *Bookshelf) NewSfacgBookshelf() {
	response := bookshelf.GET_BOOK_SHELF_INFORMATION()
	if response.Status.HTTPCode == 200 {
		if len(response.Data) == 1 {
			fmt.Println("检测到只有一个书架，无需选择书架")
			bs.ShelfIndex = 0
			bs.ShowBookshelf(response)

		} else {
			bs.ShelfIndex = tools.InputInt(">", len(response.Data))
			bs.ShowBookshelf(response)
		}
		bs.ChoiceBookshelf(response)
	} else {
		fmt.Println("get bookshelf error:", response.Status.Msg)
	}
}

func hbooker_bookshelf() (map[int][]map[string]string, error) {
	bookshelf_info := make(map[int][]map[string]string)
	shelf := HbookerAPI.GET_BOOK_SHELF_INFORMATION()

	if shelf.Code != "100000" {
		return nil, fmt.Errorf(shelf.Tip.(string))
	}
	for index, value := range shelf.Data.ShelfList {
		fmt.Println("书架号:", index, "\t\t\t书架名:", value.ShelfName)
		bookshelf_index := HbookerAPI.GET_BOOK_SHELF_INDEXES_INFORMATION(value.ShelfID)
		if bookshelf_index.Code != "100000" {
			return nil, fmt.Errorf(bookshelf_index.Tip.(string))
		}
		var bookshelf_info_list []map[string]string
		for _, book := range bookshelf_index.Data.BookList {
			bookshelf_info_list = append(bookshelf_info_list,
				map[string]string{"novel_name": book.BookInfo.BookName, "novel_id": book.BookInfo.BookID},
			)
		}
		bookshelf_info[index] = bookshelf_info_list
	}
	return bookshelf_info, nil
}

func request_bookshelf_book_list() (map[int][]map[string]string, error) {
	if config.Vars.AppType == "sfacg" {
		return sfacg_bookshelf()
	} else if config.Vars.AppType == "cat" {
		return hbooker_bookshelf()
	} else {
		return nil, fmt.Errorf("app type error")
	}
}

func InitBookShelf() ([]int, []map[string]string) {
	bookshelf_book_list, response_err := request_bookshelf_book_list()
	if response_err != nil || bookshelf_book_list == nil {
		var test_login_status bool
		fmt.Println("config.Vars.AppType ", config.Vars.AppType)
		fmt.Println("BookShelf Error:", response_err)
		if config.Vars.AppType == "sfacg" {
			test_login_status = AutoAccount()
		} else if config.Vars.AppType == "cat" {
			test_login_status = InputAccountToken()
		}
		if !test_login_status && config.Vars.AppType == "sfacg" {
			fmt.Println("please login your sfacg account and password!")
			account := tools.InputStr("please input your account:")
			password := tools.InputStr("please input your password:")
			LoginAccount(strings.TrimSpace(account), strings.TrimSpace(password), 0)
		}
		return InitBookShelf()

	}

	return select_bookcase(bookshelf_book_list)

}

func select_bookcase(bookshelf_book_list map[int][]map[string]string) ([]int, []map[string]string) {
	var bookshelf_index int
	if len(bookshelf_book_list) == 1 {
		fmt.Println("you only have one bookshelf, default loading bookshelf index:1")
		bookshelf_index = 0
	} else {
		fmt.Println("please input bookshelf index:")
		bookshelf_index = tools.InputInt(">", len(bookshelf_book_list))
	}
	book_shelf_bookcase := bookshelf_book_list[bookshelf_index]
	var bookshelf_book_index []int
	for book_index, book := range book_shelf_bookcase {
		fmt.Println("index:", book_index, "\t\tid:", book["novel_id"], "\t\tname:", book["novel_name"])
		bookshelf_book_index = append(bookshelf_book_index, book_index)
	}
	return bookshelf_book_index, book_shelf_bookcase
}
