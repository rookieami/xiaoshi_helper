package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 国际化标记
var tag = 0
var names = map[string]string{
	"xiaoshi helper": "小石助手",
	"home":           "主菜单",
}

func main() {
	a := app.New()
	iconResource, _ := fyne.LoadResourceFromPath("")
	a.SetIcon(iconResource)
	w := a.NewWindow(getName("xiaoshi helper"))
	w.SetMaster()
	showHome(w)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func makeNav(w fyne.Window) fyne.CanvasObject {
	tree := widget.NewTreeWithStrings(menuItems)
	tree.OnSelected = func(id string) {
		if id == "Request Order" {
			showUploadScreen(w)
		}
		if id == getName("home") {
			showHome(w)
		}
	}
	return container.NewBorder(nil, nil, nil, nil, tree)
}

// 创建上方工具栏
func createToolbar(w fyne.Window) *widget.Toolbar {
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			showHome(w)
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			dialog.ShowInformation("帮助", "添加微信13103965606", w)
		}),
	)
	toolbar.Theme().Color(theme.ColorYellow, 1)
	return toolbar
}
func showHome(w fyne.Window) {
	top := createToolbar(w)
	left := canvas.NewText("left", color.White)
	middle := canvas.NewText("content", color.White)
	// 创建主窗口布局，菜单栏在上方，操作区域在下方
	mainLayout := container.New(layout.NewBorderLayout(top, nil, left, nil), top, left, middle)
	w.SetContent(mainLayout)

	// 添加横线分割
	//separator := widget.NewSeparator()
	//finalLayout := container.NewVBox(toolbar, separator)
	//content := container.NewStack()
	//homeData := container.NewBorder(
	//	container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)
	//
	//split := container.NewHSplit(makeNav(w), finalLayout)
	//split.Offset = 0
}

func showUploadScreen(w fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel("Upload Excel/CSV"),
		widget.NewButton("Upload", func() {
			uploadFile(w)
		}),
	)

	split := container.NewHSplit(makeNav(w), content)
	split.Offset = 0
	w.SetContent(split)
}

func uploadFile(w fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err == nil && reader != nil {
			defer reader.Close()
			fileName := reader.URI().Name()
			ext := getFileExtension(fileName)
			if ext != ".csv" && ext != ".xls" && ext != ".xlsx" {
				dialog.ShowError(errors.New("unsupported file format"), w)
				return
			}

			// Read the content of the uploaded file
			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Convert to JSON
			jsonData, err := convertToJSON(data, ext)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Make API request
			err = makeAPIRequest(jsonData)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Optionally, keep the file for history
			// err = saveFileForHistory(fileName, data)
			// if err != nil {
			// 	log.Println("Error saving file for history:", err)
			// }

			dialog.ShowInformation("Success", "File uploaded and processed successfully", w)
		}
	}, w)
}

func getFileExtension(fileName string) string {
	return filepath.Ext(fileName)
}

func convertToJSON(data []byte, ext string) ([]byte, error) {
	if ext == ".csv" {
		return csvToJSON(data)
	} else if ext == ".xls" || ext == ".xlsx" {
		//return excelToJSON(data)
	}
	return nil, errors.New("unsupported file format")
}

func csvToJSON(data []byte) ([]byte, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return json.Marshal(records)
}

//func excelToJSON(data []byte) ([]byte, error) {
//	file, err := xlsx.OpenBinary(data)
//	if err != nil {
//		return nil, err
//	}
//
//	var rows [][]string
//	for _, sheet := range file.Sheets {
//		for _, row := range sheet.Rows {
//			var cells []string
//			for _, cell := range row.Cells {
//				cells = append(cells, cell.String())
//			}
//			rows = append(rows, cells)
//		}
//	}
//
//	return json.Marshal(rows)
//}

func makeAPIRequest(data []byte) error {
	fmt.Println("Sending data to API:", string(data))
	// Implement API request logic here
	return nil
}

func saveFileForHistory(fileName string, data []byte) error {
	// Implement logic to save the uploaded file for history
	// This is just a placeholder implementation
	filePath := filepath.Join("history", fileName)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	fmt.Println("File saved for history:", filePath)
	return nil
}

var menuItems = map[string][]string{
	"": {"Home", "Request Order"},
}

func getName(str string) string {
	if tag == 1 {
		return str
	}
	return names[str]
}
