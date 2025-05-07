package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
	"math"
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
	delEndStr = binding.NewBool()         // 删除末尾指定字符
	delAllStr = binding.NewBool()         // 删除全部指定字符
	toLower   = binding.NewBool()         // 转换为小写
	toUpper   = binding.NewBool()         // 转换为大写
	delSpace  = binding.NewBool()         // 去除空格
	strList   = make(map[string]struct{}) // 待去除字符
)

func main() {
	a := app.New()
	iconResource, _ := fyne.LoadResourceFromPath("")
	a.SetIcon(iconResource)
	w := a.NewWindow(getName("xiaoshi helper"))

	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewStack()

	setModule := func(m Module) {
		content.Objects = []fyne.CanvasObject{m.View(w)}
		content.Refresh()
	}
	modules := container.NewStack(content)
	setModule(Modules["字段处理"])
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
	return container.NewHBox(tree)
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

// ProcessExcelFile 读取 Excel 文件，处理指定列并保存为新文件
func ProcessExcelFile(filePath, outputFilePath string, oldColName, newColName string, progress *widget.ProgressBar, logOutput *widget.Entry) error {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("打开 Excel 文件失败: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// 检查 Sheet 是否存在
	sheets := f.GetSheetList()
	found := false
	for _, sheet := range sheets {
		if sheet == "Sheet1" {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("找不到工作表: Sheet1")
	}
	//读取指定列内容
	cols, err := f.GetCols("Sheet1")
	if err != nil {
		return fmt.Errorf("读取 Excel 文件失败: %v", err)
	}
	colName := ""
	var newColData []string
	for i, col := range cols {
		if len(col) == 0 || col[0] != oldColName {
			continue
		}
		colName, err = excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return fmt.Errorf("解析列名失败: %v", err)
		}
		num := len(col)
		newColData = make([]string, 0, num) // 根据实际数据长度预分配
		newColData = append(newColData, newColName)

		updateInterval := int(math.Max(1, float64(num)/100)) // 控制更新频率
		for j, str := range col {
			if j == 0 || str == "" {
				continue
			}
			newStr, err := processString(str)
			if err != nil {
				return fmt.Errorf("处理字符串失败: %v", err)
			}
			newColData = append(newColData, newStr)

			// 日志输出前判断 logOutput 是否为 nil
			if logOutput != nil {
				appendLog(logOutput, fmt.Sprintf("%d 原字符%s => %s", j, str, newStr))
			}

			// 控制进度条更新频率
			if j%updateInterval == 0 || j == num-1 {
				progress.SetValue(float64(j+1) / float64(num))
			}
		}
		break
	}
	if colName == "" {
		return fmt.Errorf("找不到列名: %s", oldColName)
	}
	// 新增一列
	err = f.InsertCols("Sheet1", colName, 1)
	if err != nil {
		return fmt.Errorf("新增列失败: %v", err)
	}

	newCol := fmt.Sprintf("%s1", colName)
	err = f.SetSheetCol("Sheet1", newCol, &newColData)
	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	// 保存文件
	err = f.SaveAs(outputFilePath)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}
	return nil
}

func processString(str string) (string, error) {
	if str == "" {
		return str, nil
	}
	// 1.移除特殊字符
	str = strings.TrimSpace(str)
	// 2.移除指定特殊字符
	if t, _ := delEndStr.Get(); t {
		// 删除末尾指定字符
		strSli := strings.Split(str, " ")
		//倒序遍历
		for i := len(strSli) - 1; i >= 0; i-- {
			// 检查当前字符是否在allStr中
			if _, ok := strList[strSli[i]]; ok {
				// 移除末尾字符
				strSli = strSli[:i]
			} else {

				break
			}
		}
		str = strings.Join(strSli, " ")
	} else if t, _ = delAllStr.Get(); t {
		// 删除所有特殊字符
		for v := range strList {
			str = strings.ReplaceAll(str, v, "")
		}
		str = strings.TrimSpace(str)
	}
	//3.转换小写
	if t, _ := toLower.Get(); t {
		str = strings.ToLower(str)
	}
	// 4.转换大写
	if t, _ := toUpper.Get(); t {
		str = strings.ToUpper(str)
	}
	// 5.删除空格
	if t, _ := delSpace.Get(); t {
		str = strings.ReplaceAll(str, " ", "")
	}
	return str, nil
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
