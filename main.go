package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
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
var (
	isRemoveSpace = false
	IsConvertCase = false
	tmpStr        = make(map[string]struct{})
)

type GlobalData struct {
	Value     string
	Listeners []func(string)
}

func (g *GlobalData) Set(value string) {
	g.Value = value
	for _, l := range g.Listeners {
		l(value)
	}
}

func (g *GlobalData) AddListener(f func(string)) {
	g.Listeners = append(g.Listeners, f)
}

var g = &GlobalData{} //全局监听
func main() {
	a := app.New()
	iconResource, _ := fyne.LoadResourceFromPath("")
	a.SetIcon(iconResource)
	w := a.NewWindow(getName("xiaoshi helper"))

	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewStack()
	title := widget.NewLabel("首页")
	intro := widget.NewLabel("添加简介")
	intro.Wrapping = fyne.TextWrapWord

	top := container.NewVBox(title, widget.NewSeparator(), intro)
	setModule := func(m Module) {
		title.SetText(m.Title)
		isMarkdown := len(m.Intro) == 0
		if !isMarkdown {
			intro.SetText(m.Intro)
		}
		if m.Title == "Welcome" || isMarkdown {
			top.Hide()
		} else {
			top.Show()
		}

		content.Objects = []fyne.CanvasObject{m.View(w)}
		content.Refresh()
	}
	modules := container.NewBorder(top, nil, nil, nil, content)
	split := container.NewHSplit(makeNav(setModule), modules)
	split.Offset = 0.2
	w.SetContent(split)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func makeNav(setModule func(m Module)) fyne.CanvasObject {
	a := fyne.CurrentApp()
	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return moduleIndexs[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := moduleIndexs[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := Modules[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if t, ok := Modules[uid]; ok {
				//for _, f := range tutorials.OnChangeFuncs {
				//	f()
				//}
				//tutorials.OnChangeFuncs = nil // Loading a page registers a new cleanup.

				a.Preferences().SetString("", uid)
				setModule(t)
			}
		},
	}
	return container.NewBorder(nil, nil, nil, nil, tree)
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

	// 创建主题选项
	themeMenu := createThemeMenus(a, w)

	// 组装主菜单
	mainMenu := fyne.NewMainMenu(
		m,
		fyne.NewMenu("help", helpMenuItem),
		themeMenu,
	)
	return mainMenu
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

// 创建主题菜单
func createThemeMenus(a fyne.App, w fyne.Window) *fyne.Menu {
	// 输入参数校验
	if a == nil {
		return nil // 返回 nil 表示无法创建菜单
	}

	// 定义主题设置的辅助函数
	setTheme := func(variant fyne.ThemeVariant) {
		defaultTheme := theme.DefaultTheme()
		defaultTheme.Color(theme.ColorNameBackground, variant)
		a.Settings().SetTheme(defaultTheme)
		// 刷新窗口以应用新主题
		w.Canvas().Refresh(w.Content())
	}

	// 创建主题菜单项
	themeMenuItem := fyne.NewMenuItem("跟随系统", func() {
		a.Settings().SetTheme(theme.DefaultTheme())
	})

	lightThemeItem := fyne.NewMenuItem("浅色", func() {
		setTheme(theme.VariantLight)
	})

	darkThemeItem := fyne.NewMenuItem("深色", func() {
		a.Settings().SetTheme(&ForcedVariant{Theme: theme.DefaultTheme(), ThemeVariant: theme.VariantDark})
	})

	// 返回完整的主题菜单
	return fyne.NewMenu("主题", themeMenuItem, lightThemeItem, darkThemeItem)
}

func selectFiled(w fyne.Window) fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewButton("字符替换", func() {
			uploadFile(w)
		}),
	)
	return content
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

			// 读取文件
			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			//
			jsonData, err := convertToJSON(data, ext)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// 保存文件
			err = saveFileForHistory(fileName, jsonData)
			if err != nil {
				log.Println("Error saving file for history:", err)
			}

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

//var menuItems = map[string][]string{
//	"": {"Home", "Request Order"},
//}

func getName(str string) string {
	if tag == 1 {
		return str
	}
	return names[str]
}

// 从配置或环境变量中加载联系信息
func getContactInfo(key string) string {
	// 示例：从环境变量中获取联系信息
	// 实际实现可以根据需求调整为读取配置文件等
	return os.Getenv(key)
}
