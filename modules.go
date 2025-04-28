package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	// 创建一个可以选择指定文件的组件
	labal := widget.NewLabel("文件名称: ")
	button := widget.NewButton("选择文件", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				defer reader.Close()
				fileName := reader.URI().Name()
				ext := getFileExtension(fileName)
				if ext != ".csv" && ext != ".xls" && ext != ".xlsx" {
					dialog.ShowError(errors.New("unsupported file format"), w)
					return
				}
				labal.SetText(fmt.Sprintf("文件路径:  %s", fileName))
			}
		}, w)
	})
	input := widget.NewMultiLineEntry()
	selectRow := container.NewHBox(
		widget.NewLabel("输入待替换字段名"),
		input,
	)
	selectBox := container.NewHBox(
		widget.NewLabel("选择操作"),
		widget.NewCheckGroup([]string{"去除空格", "转换大小写"}, func(values []string) {
			if len(values) > 0 {
				for _, value := range values {
					if value == "去除空格" {
						isRemoveSpace = true
					}
					if value == "转换大小写" {
						IsConvertCase = true
					}
				}
			}
		}),
	)
	// 创建一个容器,其中有两个按钮，点击左侧按钮后，上方新增一个输入框
	inputList := container.NewVBox()
	content2 := container.NewHBox(
		//新建两个按钮，
		widget.NewButton("新增", func() {
			//点击后，inputList新增一个输入框
			input := widget.NewEntry()
			input.SetPlaceHolder("")
			if input.Text != "" {
				tmpStr[input.Text] = struct{}{}
			}
			inputList.Add(input)
		}),
	)
	content1 := container.NewVBox(inputList, content2)
	content := container.NewVBox(button, labal, selectRow, selectBox, content1)
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
