package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_demo/tutorials"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

const preferenceCurrentTutorial = "currentTutorial"

var topWindow fyne.Window
var HasNativeMenu = true

func main() {
	a := app.New()
	iconResource, _ := fyne.LoadResourceFromPath("")
	a.SetIcon(iconResource)
	w := a.NewWindow(getName("xiaoshi helper"))
	topWindow = w

	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewStack()

	//title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = fyne.TextWrapWord

	//top := container.NewVBox(title, widget.NewSeparator(), intro)
	//setTutorial := func(t tutorials.Tutorial) {
	//if fyne.CurrentDevice().IsMobile() {
	//	child := a.NewWindow(t.Title)
	//	topWindow = child
	//	child.SetContent(t.View(topWindow))
	//	child.Show()
	//	child.SetOnClosed(func() {
	//		topWindow = w
	//	})
	//	return
	//}
	//
	//title.SetText(t.Title)
	//isMarkdown := len(t.Intro) == 0
	//if !isMarkdown {
	//	intro.SetText(t.Intro)
	//}
	//
	//if t.Title == "Welcome" || isMarkdown {
	//	top.Hide()
	//} else {
	//	top.Show()
	//}
	//
	//content.Objects = []fyne.CanvasObject{t.View(w)}
	//content.Refresh()
	//}

	//tutorial := container.NewBorder(top, nil, nil, nil, content)

	split := container.NewHSplit(makeNav(w), content)
	split.Offset = 0
	w.SetContent(split)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}
func makeMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	const contactInfoKey = "contact_info"
	contactInfo := getContactInfo(contactInfoKey) // 从配置或环境变量中加载联系信息

	//创建主菜单
	m := fyne.NewMenu("todo",
		fyne.NewMenuItem("Home", func() {
			// Handle Home menu item click
		}),
	)

	// 创建帮助菜单项
	helpMenuItem := createHelpMenuItem(w, contactInfo)

	// 组装主菜单
	mainMenu := fyne.NewMainMenu(
		m,
		fyne.NewMenu("help", helpMenuItem),
	)
	return mainMenu
}
func makeNav(w fyne.Window) fyne.CanvasObject {
	tree := widget.NewTreeWithStrings(menuItems)
	tree.OnSelected = func(id string) {
		if id == "Request Order" {
			//showUploadScreen(w)
		}
		if id == getName("home") {
			//showHome(w)
		}
	}
	return container.NewBorder(nil, nil, nil, nil, tree)
}
func makeNav1(setTutorial func(tutorial tutorials.Tutorial), loadPrevious bool) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return tutorials.TutorialIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := tutorials.TutorialIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := tutorials.Tutorials[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if t, ok := tutorials.Tutorials[uid]; ok {
				for _, f := range tutorials.OnChangeFuncs {
					f()
				}
				tutorials.OnChangeFuncs = nil // Loading a page registers a new cleanup.

				a.Preferences().SetString(preferenceCurrentTutorial, uid)
				setTutorial(t)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentTutorial, "welcome")
		tree.Select(currentPref)
	}

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantLight})
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
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

type forcedVariant struct {
	fyne.Theme

	variant fyne.ThemeVariant
}

func (f *forcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

// 从配置或环境变量中加载联系信息
func getContactInfo(key string) string {
	// 示例：从环境变量中获取联系信息
	// 实际实现可以根据需求调整为读取配置文件等
	return os.Getenv(key)
}

// 创建帮助菜单项
func createHelpMenuItem(w fyne.Window, contactInfo string) *fyne.MenuItem {
	if contactInfo == "" {
		contactInfo = "联系信息未配置" // 默认值，避免空字符串
	}
	return fyne.NewMenuItem("联系我们", func() {
		// 异常处理
		defer func() {
			if r := recover(); r != nil {
				dialog.ShowError(fmt.Errorf("显示信息时发生错误: %v", r), w)
			}
		}()
		dialog.ShowInformation("", contactInfo, w)
	})
}
