package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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
	"字段替换": {"字段替换", "excel表替换指定字段", ReplaceCharWindow},
	"todo": {"todo2", "to", Todo},
}

var moduleIndexs = map[string][]string{
	"": {"字段替换", "todo"},
	//"替换字符": {"替换字符", "添加注释"},
}

// 替换字符窗口
func ReplaceCharWindow(w fyne.Window) fyne.CanvasObject {
	path := ""
	labal := widget.NewLabel("已选文件: ")
	// 创建一个可以选择指定文件的组件
	button := widget.NewButton("选择文件", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				defer reader.Close()
				fileName := reader.URI().Name()
				ext := getFileExtension(fileName)
				if ext != ".csv" && ext != ".xls" && ext != ".xlsx" {
					dialog.ShowError(errors.New("选择文件格式错误"), w)
					return
				}
				path = reader.URI().Path()
				labal.SetText(fmt.Sprintf("已选文件: %s", path))
			}
		}, w)
	})

	topObj := container.New(layout.NewHBoxLayout(), button, layout.NewSpacer(), labal, layout.NewSpacer())

	label1 := widget.NewLabel("待操作字段名")
	entry1 := widget.NewEntry()
	label2 := widget.NewLabel("输出新字段名")
	entry2 := widget.NewEntry()
	nameObj := container.New(layout.NewGridLayout(4), label1, entry1, label2, entry2)

	// 创建一个容器,其中有两个按钮，点击左侧按钮后，上方新增一个输入框
	inputList := container.NewGridWithColumns(2)
	button2 := widget.NewButton("新增指定去除字符", func() {
		//点击后，inputList新增一个输入框
		input := widget.NewEntry()
		if input.Text != "" {
			tmpStr[input.Text] = struct{}{}
		}
		inputList.Add(layout.NewSpacer())
		inputList.Add(input)
	})
	content1 := container.NewGridWithColumns(2, button2, layout.NewSpacer(), layout.NewSpacer(), inputList)

	label3 := widget.NewLabel("操作选项")
	checkGroup := widget.NewCheckGroup([]string{"去除末尾", "去除全部", "转换小写", "转换大写", "去除空格"}, func(selected []string) {
		for _, v := range selected {
			tmpStr[v] = struct{}{}
		}
	})
	content3 := container.NewVBox(label3, checkGroup)

	content := container.NewVBox(topObj, layout.NewSpacer(), nameObj, content1, content3)
	return content
}
func Todo(w fyne.Window) fyne.CanvasObject {
	content := container.NewVBox(
	//widget.NewButton("字符替换", func() {
	//	uploadFile(w)
	//}),
	)
	return content
}
