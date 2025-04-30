package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"path/filepath"
)

// OnChangeFuncs is a slice of functions that can be registered
// to run when the user switches tutorial.
var OnChangeFuncs []func()

// 组件
type Module struct {
	Title string
	Intro string
	View  func(w fyne.Window) fyne.CanvasObject
}

var Modules = map[string]Module{
	"字段处理": {"处理表指定列数据", "", ReplaceCharWindow},
	"todo": {"todo2", "to", Todo},
}

var moduleIndexs = map[string][]string{
	"": {"字段处理", "todo"},
	//"替换字符": {"替换字符", "添加注释"},
}

// 替换字符窗口
func ReplaceCharWindow(w fyne.Window) fyne.CanvasObject {
	path := ""
	labal := widget.NewLabel("选择文件: ")
	// 创建一个可以选择指定文件的组件
	button := widget.NewButton("选择文件", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				dialog.ShowError(errors.New("选择文件错误"), w)
				return
			}
			defer reader.Close()

			fileName := reader.URI().Name()
			ext := filepath.Ext(fileName)
			//if ext != ".csv" && ext != ".xls" && ext != ".xlsx"
			if ext != ".xls" && ext != ".xlsx" {
				dialog.ShowError(errors.New("文件格式错误"), w)
				return
			}
			path = reader.URI().Path()
			labal.SetText(fmt.Sprintf("已选文件:  %s", path))
		}, w)
	})

	topObj := container.NewHBox(labal, layout.NewSpacer(), button, layout.NewSpacer())

	label1 := widget.NewLabel("操作字段名")
	column := widget.NewEntry()
	label2 := widget.NewLabel("输出新字段名")
	newColumn := widget.NewEntry()
	nameObj := container.New(layout.NewGridLayout(4), label1, column, label2, newColumn)

	label4 := widget.NewLabel("操作选项:")
	check1 := widget.NewCheck("删除末尾指定字符", func(b bool) {

	})
	check2 := widget.NewCheck("删除全部指定字符", func(b bool) {
	})
	check3 := widget.NewCheck("转换小写", func(b bool) {
	})
	check4 := widget.NewCheck("转换大写", func(b bool) {
	})
	check5 := widget.NewCheck("去除空格", func(b bool) {
	})
	content3 := container.NewGridWithColumns(4, label4, check1, check2, check3, layout.NewSpacer(), check4, check5)

	input3 := widget.NewEntry()
	button2 := widget.NewButton("追加删除字符", func() {
	})
	content1 := container.NewGridWithColumns(4, button2, input3)
	button2.OnTapped = func() {
		input := widget.NewEntry()
		content1.Add(input)
	}
	progress := widget.NewProgressBar()
	progress.Hide()

	//新增一个日志输出框，动态显示打印结果
	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord // 自动换行
	logOutput.Disable()                    // 设置为只读
	// 将输入框放入滚动容器
	button3 := widget.NewButton("开始执行", func() {
		if path == "" {
			dialog.ShowError(errors.New("请选择文件"), w)
			return
		}
		// 读取excel表
		//data := readExcel(path)
		if column.Text == "" {
			dialog.ShowError(errors.New("请输入操作字段名"), w)
			return
		}
		//if newColumn.Text == "" {
		//	dialog.ShowError(errors.New("请输入新字段名"), w)
		//	return
		//}
		// 读取excel表
		if logOutput.Disabled() {
			logOutput.Enable()
		}
		appendLog(logOutput, "开始执行...")
		appendLog(logOutput, fmt.Sprintf("文件路径 %s", path))

		progress.Show()
		for i := 1; i <= 100; i++ {
			progress.SetValue(float64(i) / 100)
			appendLog(logOutput, fmt.Sprintf("进度%d", i))
		}
		// 操作完成后隐藏进度条
	})
	content := container.NewVBox(topObj, nameObj, content3, content1, button3, progress)
	//新建一个二分布局
	content5 := container.NewGridWithRows(2, content, logOutput)
	return content5
}
func Todo(w fyne.Window) fyne.CanvasObject {
	content := container.NewBorder(layout.NewSpacer(), widget.NewLabel("todo"), nil, nil, nil)
	return content
}

// appendLog 向日志框追加内容，并自动滚动到底部
func appendLog(log *widget.Entry, text string) {
	if text == "" {
		return // 避免无意义操作
	}
	// 追加到最上方
	log.SetText(text + "\n" + log.Text)
}
