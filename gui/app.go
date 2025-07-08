package gui

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	_ "github.com/rs/zerolog/log"
	"net/url"
	_ "os"
	"sort"
	"strconv"
	"strings"
	"time"

	"Cursor_Windsurf_Reset/cleaner"
	"Cursor_Windsurf_Reset/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// App represents the main GUI application
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	engine     *cleaner.Engine
	config     *config.Config
	logChan    chan string

	guiLogger zerolog.Logger // 新增GUI专用logger

	// UI components
	appData           []AppInfo
	progressBar       *widget.ProgressBar
	statusLabel       *widget.Label
	logText           *widget.Entry
	cleanButton       *widget.Button
	discoverButton    *widget.Button
	configButton      *widget.Button
	aboutButton       *widget.Button
	helpButton        *widget.Button
	selectedIndex     int
	mainAreaContainer *fyne.Container

	selectedApps   map[int]bool
	selectAllCheck *widget.Check
}

// AppInfo represents application information for the UI
type AppInfo struct {
	Name        string
	DisplayName string
	Path        string
	Size        string
	Running     bool
	Found       bool
	Selected    bool // 新增选中状态字段
}

// NewApp creates a new GUI application
func NewApp() *App {
	fyneApp := app.New()
	fyneApp.SetIcon(theme.ComputerIcon())

	fyneApp.Settings().SetTheme(NewModernDarkTheme())

	// Setup zerolog
	logChan := make(chan string, 100)
	guiWriter := &config.GuiLogWriter{LogChan: logChan}
	// config.SetupLogger(guiWriter) // 移除全局log.Logger覆盖

	// 为GUI日志单独创建ConsoleWriter，输出纯文本LEVEL: message格式
	consoleWriter := zerolog.ConsoleWriter{
		Out:             guiWriter,
		NoColor:         true,
		TimeFormat:      "",
		FormatTimestamp: func(i interface{}) string { return "" },
		FormatLevel: func(i interface{}) string {
			if l, ok := i.(string); ok {
				return strings.ToUpper(l) + ":"
			}
			return "INFO:"
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf(" %s", i)
		},
	}
	guiLogger := zerolog.New(consoleWriter).With().Logger()

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		guiLogger.Error().Err(err).Msg("Failed to load configuration")
		cfg = config.GetDefaultConfig()
	}

	// Create cleaning engine
	engine := cleaner.NewEngine(cfg, false, false)

	app := &App{
		fyneApp:       fyneApp,
		engine:        engine,
		config:        cfg,
		logChan:       logChan,
		guiLogger:     guiLogger, // 赋值
		selectedApps:  make(map[int]bool),
		selectedIndex: -1, // 初始化为-1表示未选中
	}

	app.setupMainWindow()
	go app.listenForLogs()
	return app
}

func (app *App) listenForLogs() {
	for logMsg := range app.logChan {
		currentText := app.logText.Text
		if len(currentText) > 20000 { // Increased limit
			lines := strings.Split(currentText, "\n")
			if len(lines) > 400 {
				currentText = strings.Join(lines[len(lines)-300:], "\n")
			}
		}
		app.logText.SetText(currentText + logMsg)
		app.logText.CursorRow = len(strings.Split(app.logText.Text, "\n"))
	}
}

// setupMainWindow sets up the main application window
func (app *App) setupMainWindow() {
	app.mainWindow = app.fyneApp.NewWindow("Cursor & Windsurf 数据重置工具")
	app.mainWindow.Resize(fyne.NewSize(860, 600)) // 调整为更紧凑的高度
	app.mainWindow.CenterOnScreen()
	app.mainWindow.SetIcon(theme.ComputerIcon())
	app.mainWindow.SetMaster()

	// 设置固定最小窗口大小，确保UI元素不会挤压变形
	app.mainWindow.SetFixedSize(false)
	// Fyne不支持SetMinSize，使用Resize代替
	app.mainWindow.Resize(fyne.NewSize(750, 550))

	// Create UI components
	app.createUI()

	// Set up event handlers
	app.setupEventHandlers()

	// 延迟执行初始扫描，等待UI完全初始化
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Initial discovery
		app.performDiscovery()
	}()
}

// ModernButton 创建一个带有悬停效果和更现代外观的按钮
func ModernButton(text string, icon fyne.Resource, onTapped func()) *widget.Button {
	button := widget.NewButtonWithIcon(text, icon, onTapped)

	// 设置按钮重要性为高，使其有更明显的视觉效果
	button.Importance = widget.MediumImportance

	return button
}

// createUI creates the main UI layout
func (app *App) createUI() {
	// 头部区域 - 使用垂直布局添加图标和标题
	// 初始化帮助和关于按钮
	app.helpButton = ModernButton("", theme.HelpIcon(), app.onHelp)
	app.aboutButton = ModernButton("", theme.InfoIcon(), app.onAbout)

	// 创建应用标题，增加大小和样式
	appTitle := widget.NewLabelWithStyle(
		"Cursor & Windsurf 数据重置工具",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// 如果需要更大的标题，可以创建一个更大的标签
	// 使用主题的标题大小

	// 美化头部布局，增加头部间距
	header := container.NewVBox(
		container.NewPadded(
			container.NewHBox(
				widget.NewIcon(theme.ComputerIcon()),
				appTitle,
				layout.NewSpacer(),
				app.helpButton,
				app.aboutButton,
			),
		),
		widget.NewSeparator(),
	)

	// 应用列表区域 - 使用卡片容器增加视觉层次感
	listLabel := widget.NewLabelWithStyle("应用程序列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// 创建加载指示器和提示文本的组合，增加动态效果
	loadingLabel := widget.NewLabelWithStyle("正在加载应用列表，请稍候...",
		fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	loadingIcon := widget.NewIcon(theme.ViewRefreshIcon())

	// 包装在一个卡片容器中，增强视觉层次感
	loadingContainer := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), loadingIcon, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), loadingLabel, layout.NewSpacer()),
	)

	// 使用Padding增加美观度，添加边框和阴影效果
	appListContainer := container.NewBorder(listLabel, nil, nil, nil,
		container.NewPadded(loadingContainer))

	// 操作区域
	app.progressBar = widget.NewProgressBar()
	app.progressBar.Hide()

	// 状态标签不再直接显示在界面上，但仍然保留用于日志记录
	app.statusLabel = widget.NewLabel("就绪")
	app.statusLabel.Hide()

	// 操作按钮区域 - 使用卡片布局提高视觉层次感
	// 初始化按钮，使用更明亮的图标和悬停效果
	app.discoverButton = ModernButton("扫描应用", theme.SearchIcon(), app.onDiscover)
	app.cleanButton = ModernButton("重置选中", theme.DeleteIcon(), app.onClean)
	app.configButton = ModernButton("设置", theme.SettingsIcon(), app.onConfig)

	// 设置按钮重要性级别
	app.discoverButton.Importance = widget.HighImportance
	app.cleanButton.Importance = widget.DangerImportance
	app.configButton.Importance = widget.MediumImportance

	// 禁用重置按钮，直到选中应用
	app.cleanButton.Disable()

	// 创建卡片式操作按钮区域
	actionButtonsCard := container.NewVBox(
		widget.NewLabelWithStyle("操作", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(3,
			app.discoverButton,
			app.cleanButton,
			app.configButton,
		),
	)

	// 初始化全选复选框
	app.selectAllCheck = widget.NewCheck("全选", func(checked bool) {
		app.log("INFO", fmt.Sprintf("全选状态变更: %v", checked))

		// 重置选中状态
		app.selectedApps = make(map[int]bool)

		// 更新每个应用的选中状态
		for i, appInfo := range app.appData {
			if appInfo.Found && !appInfo.Running {
				app.selectedApps[i] = checked
			}
		}

		// 重新创建应用列表
		app.refreshAppList()

		// 更新重置按钮状态
		app.updateCleanButton()
	})

	// 状态区域 - 只保留进度条
	statusCard := container.NewVBox(
		widget.NewLabelWithStyle("进度", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		app.progressBar,
	)

	// 功能区组合 - 添加卡片式背景
	controlsContainer := container.NewVBox(
		actionButtonsCard, // 移除额外的内边距
		widget.NewSeparator(),
		statusCard, // 移除额外的内边距
	)

	// 日志区域 - 减小高度使其更紧凑
	app.logText = widget.NewMultiLineEntry()
	app.logText.Disable()
	app.logText.SetPlaceHolder("操作日志将显示在此处...")
	app.logText.TextStyle = fyne.TextStyle{Monospace: true}

	// 创建折叠按钮
	var collapseLogButton *widget.Button
	var clearLogButton *widget.Button
	var logContentContainer *fyne.Container

	// 先初始化按钮
	collapseLogButton = ModernButton("", theme.MoveDownIcon(), nil)
	clearLogButton = ModernButton("清除日志", theme.ContentClearIcon(), func() {
		app.logText.SetText("")
	})

	// 创建日志容器
	logScrollContainer := container.NewScroll(app.logText)
	// 增加日志显示区域的高度
	logScrollContainer.SetMinSize(fyne.NewSize(0, 150))

	// 创建一个变量引用日志内容容器
	logContentContainer = container.NewVBox(
		widget.NewSeparator(),
		logScrollContainer,
	)

	// 创建日志折叠状态变量
	isLogCollapsed := false

	// 设置折叠按钮的回调函数
	collapseLogButton.OnTapped = func() {
		isLogCollapsed = !isLogCollapsed

		if isLogCollapsed {
			// 折叠状态
			collapseLogButton.SetIcon(theme.MoveUpIcon())
			logContentContainer.Hide()
		} else {
			// 展开状态
			collapseLogButton.SetIcon(theme.MoveDownIcon())
			logContentContainer.Show()
		}
	}

	// 创建日志标题区域
	logTitle := container.NewHBox(
		widget.NewLabelWithStyle("操作日志", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		clearLogButton,
		collapseLogButton,
	)

	// 完整日志容器
	logContainer := container.NewBorder(
		logTitle,
		nil, nil, nil,
		logContentContainer,
	)

	// 创建边框和阴影效果的函数
	createBorderedContainer := func(content fyne.CanvasObject) *fyne.Container {
		border := canvas.NewRectangle(color.NRGBA{R: 50, G: 55, B: 65, A: 100})
		border.StrokeWidth = 1
		border.StrokeColor = color.NRGBA{R: 60, G: 70, B: 80, A: 150}

		return container.New(
			layout.NewMaxLayout(),
			border,
			content,
		)
	}

	// 创建控制区域的边框容器
	borderedControlsContainer := createBorderedContainer(controlsContainer)

	// 创建日志区域的边框容器
	borderedLogContainer := createBorderedContainer(logContainer)

	// 将应用列表和控制区域放在一起
	app.mainAreaContainer = container.NewBorder(
		nil,
		borderedControlsContainer, // 使用带边框的控制区域
		nil, nil,
		appListContainer,
	)

	// 主容器 - 将日志区域放在下方，调整边距使更紧凑
	mainContent := container.NewBorder(
		header,
		borderedLogContainer, // 使用带边框的日志区域
		nil, nil,
		app.mainAreaContainer,
	)

	// 设置更小的内边距，提高紧凑性
	paddedContent := container.NewPadded(mainContent)

	app.mainWindow.SetContent(paddedContent)

	// 调整窗口大小
	app.mainWindow.Resize(fyne.NewSize(860, 600)) // 减小高度使界面更紧凑
}

// setupEventHandlers sets up event handlers for the UI
func (app *App) setupEventHandlers() {
	// 事件处理器已经在createUI方法中设置
	// 如果有其他事件处理器，可以在这里添加
}

// performDiscovery performs application discovery
func (app *App) performDiscovery() {
	app.log("INFO", "开始发现应用程序...")
	app.statusLabel.SetText("正在发现应用程序...")

	// 获取和显示所有应用数据路径
	appDataPaths := app.engine.GetAppDataPaths()

	// 打印原始路径数据
	app.log("INFO", fmt.Sprintf("原始应用数据路径: %v", appDataPaths))

	// 重置应用数据列表
	app.appData = make([]AppInfo, 0)

	// 调试日志
	app.log("INFO", fmt.Sprintf("发现应用数量: %d", len(appDataPaths)))

	// 详细输出所有应用
	for name, path := range appDataPaths {
		app.log("INFO", fmt.Sprintf("发现应用: %s, 路径: %s", name, path))
	}

	// 按顺序排列应用，确保顺序一致
	appNames := make([]string, 0, len(appDataPaths))
	for appName := range appDataPaths {
		appNames = append(appNames, appName)
	}
	// 按应用名称排序，保证顺序一致
	sort.Strings(appNames)

	// 按排序后的顺序处理应用
	for _, appName := range appNames {
		appPath := appDataPaths[appName]
		appConfig := app.config.Applications[appName]

		app.log("INFO", fmt.Sprintf("处理应用: %s, 显示名称: %s", appName, appConfig.DisplayName))

		appInfo := AppInfo{
			Name:        appName,
			DisplayName: appConfig.DisplayName,
			Path:        appPath,
			Found:       appPath != "",
			Selected:    false, // 确保初始未选中
		}

		if appInfo.Found {
			// 检查应用是否正在运行
			appInfo.Running = app.engine.IsAppRunning(appName)

			// 获取目录大小
			size := app.engine.GetDirectorySize(appPath)
			appInfo.Size = app.engine.FormatSize(size)

			app.log("INFO", fmt.Sprintf("发现应用详情: 显示名称: %s, 路径: %s, 大小: %s, 运行中: %v", appInfo.DisplayName, appPath, appInfo.Size, appInfo.Running))
		} else {
			appInfo.Size = "未知"
			app.log("INFO", fmt.Sprintf("未找到应用: %s", appInfo.DisplayName))
		}

		app.appData = append(app.appData, appInfo)
		app.log("INFO", fmt.Sprintf("添加应用到列表: 显示名称: %s, 索引: %d", appInfo.DisplayName, len(app.appData)-1))
	}

	// 调试日志
	app.log("INFO", fmt.Sprintf("列表中应用总数: %d", len(app.appData)))
	for i, appInfo := range app.appData {
		app.log("INFO", fmt.Sprintf("最终应用列表项: 索引: %d, 显示名称: %s, 路径: %s", i, appInfo.DisplayName, appInfo.Path))
	}

	// 清空选中状态
	app.selectedApps = make(map[int]bool)

	// 安全地设置全选复选框状态
	if app.selectAllCheck != nil {
		app.selectAllCheck.SetChecked(false)
	}

	// 重新创建并刷新应用列表
	app.refreshAppList()

	// 确保在主UI线程上执行刷新
	fyne.CurrentApp().Driver().CanvasForObject(app.mainWindow.Content()).Refresh(app.mainWindow.Content())

	app.statusLabel.SetText("发现完成")
	app.log("INFO", "应用程序发现已完成")

	// 计算有效的应用数量（已找到且未运行的应用）
	validAppCount := 0
	for _, appInfo := range app.appData {
		if appInfo.Found && !appInfo.Running {
			validAppCount++
		}
	}

	// 在日志中额外添加摘要信息
	app.log("INFO", fmt.Sprintf("共发现 %d 个应用，其中 %d 个可重置", len(app.appData), validAppCount))

	// 更新重置按钮状态
	app.updateCleanButton()
}

// onDiscover handles the discover button click
func (app *App) onDiscover() {
	// 禁用扫描按钮，避免重复点击
	app.discoverButton.Disable()
	app.discoverButton.SetText("正在扫描...")

	// 显示加载提示
	app.log("INFO", "用户启动了发现流程")

	// 在后台执行扫描，避免UI卡顿
	go func() {
		// 执行发现过程
		app.performDiscovery()

		// 操作完成后，恢复按钮状态
		app.discoverButton.SetText("扫描应用")
		app.discoverButton.Enable()

		// 确保UI在主线程上刷新
		if canvas := fyne.CurrentApp().Driver().CanvasForObject(app.mainWindow.Content()); canvas != nil {
			canvas.Refresh(app.mainWindow.Content())
		}
	}()
}

// updateCleanButton 更新重置按钮状态
func (app *App) updateCleanButton() {
	// 检查是否有选中的应用
	hasSelected := false

	// 调试日志 - 输出所有应用信息
	app.log("INFO", "更新重置按钮状态")
	for i, appInfo := range app.appData {
		isSelected := app.selectedApps[i]
		app.log("INFO", fmt.Sprintf("应用状态用于按钮更新: 索引: %d, 显示名称: %s, 已找到: %v, 运行中: %v, 选中: %v", i, appInfo.DisplayName, appInfo.Found, appInfo.Running, isSelected))
	}

	for _, selected := range app.selectedApps {
		if selected {
			hasSelected = true
			break
		}
	}

	// 根据是否有选中的应用启用或禁用重置按钮
	if hasSelected {
		app.cleanButton.Enable()
		// 计算选中的数量
		count := 0
		for _, selected := range app.selectedApps {
			if selected {
				count++
			}
		}
		app.cleanButton.SetText(fmt.Sprintf("重置选中 (%d)", count))
		app.log("INFO", fmt.Sprintf("重置按钮启用, 选中数量: %d", count))
	} else {
		app.cleanButton.Disable()
		app.cleanButton.SetText("重置选中")
		app.log("INFO", "重置按钮禁用")
	}
}

// onClean handles the clean button click
func (app *App) onClean() {
	// 获取所有选中的有效应用
	selectedApps := make([]AppInfo, 0)
	for id, selected := range app.selectedApps {
		if selected && id < len(app.appData) {
			appInfo := app.appData[id]
			if appInfo.Found && !appInfo.Running {
				selectedApps = append(selectedApps, appInfo)
			}
		}
	}

	// 如果没有选中应用，直接返回
	if len(selectedApps) == 0 {
		dialog.ShowInformation("提示", "请先选择要重置的应用", app.mainWindow)
		return
	}

	// 检查是否有应用正在运行
	for _, appInfo := range selectedApps {
		if appInfo.Running {
			dialog.ShowError(fmt.Errorf("请先关闭 %s 再重置", appInfo.DisplayName), app.mainWindow)
			return
		}
	}

	// 创建确认内容
	confirmContent := container.NewVBox(
		widget.NewLabelWithStyle(
			fmt.Sprintf("您即将重置 %d 个应用的数据", len(selectedApps)),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		widget.NewSeparator(),
	)

	// 添加选中的应用名称
	for _, appInfo := range selectedApps {
		confirmContent.Add(widget.NewLabel("• " + appInfo.DisplayName))
	}

	// 添加操作说明
	confirmContent.Add(widget.NewSeparator())
	confirmContent.Add(widget.NewLabel("此操作将会："))
	confirmContent.Add(widget.NewLabel("• 重置设备ID与唯一标识"))
	confirmContent.Add(widget.NewLabel("• 清除账户登录记录与凭据"))
	confirmContent.Add(widget.NewLabel("• 删除缓存数据与历史记录"))
	confirmContent.Add(widget.NewLabel("• 创建所有修改文件的备份"))
	confirmContent.Add(widget.NewSeparator())
	confirmContent.Add(widget.NewLabelWithStyle(
		"备份将保存在您的主文件夹中",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	))

	// 显示确认对话框
	customConfirm := dialog.NewCustomConfirm(
		"确认重置操作",
		"确认执行",
		"取消",
		confirmContent,
		func(confirm bool) {
			if confirm {
				// 逐个重置选中的应用
				for _, appInfo := range selectedApps {
					app.performCleanup(appInfo)
				}
			}
		},
		app.mainWindow,
	)

	customConfirm.Show()
}

// performCleanup performs the actual cleanup operation
func (app *App) performCleanup(appInfo AppInfo) {
	app.log("INFO", fmt.Sprintf("开始重置: %s", appInfo.DisplayName))
	app.statusLabel.SetText(fmt.Sprintf("正在重置: %s", appInfo.DisplayName))
	app.progressBar.Show()
	app.progressBar.SetValue(0)

	// Update engine settings
	app.engine = cleaner.NewEngine(app.config, false, false)

	// Start progress monitoring
	go app.monitorProgress()

	// Perform cleanup in background
	go func() {
		err := app.engine.CleanApplication(context.Background(), appInfo.Name)
		if err != nil {
			app.log("ERROR", fmt.Sprintf("重置失败: %s, 错误: %v", appInfo.DisplayName, err))
		} else {
			app.log("INFO", fmt.Sprintf("重置完成: %s", appInfo.DisplayName))
		}
	}()
}

// monitorProgress monitors cleanup progress
func (app *App) monitorProgress() {
	progressChan := app.engine.GetProgressChannel()
	for update := range progressChan {
		app.progressBar.SetValue(update.Progress / 100.0)
		app.statusLabel.SetText(update.Message)
		app.log("INFO", fmt.Sprintf("阶段: %s, 进度: %s, 百分比: %.2f%%", update.Phase, update.Message, update.Progress))
	}
}

// onConfig handles the config button click
func (app *App) onConfig() {
	// 创建配置对话框
	configForm := &widget.Form{}

	// 备份设置
	backupEnabledCheck := widget.NewCheck("启用备份功能", nil)
	backupEnabledCheck.SetChecked(app.config.BackupOptions.Enabled)

	backupKeepDays := widget.NewEntry()
	backupKeepDays.SetText(fmt.Sprintf("%d", app.config.BackupOptions.RetentionDays))

	// 安全设置
	confirmCheck := widget.NewCheck("操作需要确认", nil)
	confirmCheck.SetChecked(app.config.SafetyOptions.RequireConfirmation)

	// 添加到表单
	configForm.Append("启用备份", backupEnabledCheck)
	configForm.Append("保留天数", backupKeepDays)
	configForm.Append("需要确认", confirmCheck)

	// 创建对话框
	dialog.ShowCustomConfirm("应用设置", "保存", "取消", configForm, func(save bool) {
		if save {
			// 更新配置
			app.config.BackupOptions.Enabled = backupEnabledCheck.Checked
			days, err := strconv.Atoi(backupKeepDays.Text)
			if err == nil && days > 0 {
				app.config.BackupOptions.RetentionDays = days
			}
			app.config.SafetyOptions.RequireConfirmation = confirmCheck.Checked

			err = config.SaveConfig(app.config, "")
			if err != nil {
				dialog.ShowError(fmt.Errorf("保存配置失败: %v", err), app.mainWindow)
				app.log("ERROR", fmt.Sprintf("保存配置失败: %v", err))
			} else {
				app.log("INFO", "配置已保存")
			}
		}
	}, app.mainWindow)
}

// onHelp handles the help button click
func (app *App) onHelp() {
	helpContent := container.NewVBox(
		widget.NewLabelWithStyle("使用说明", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("1. 点击【扫描应用】按钮检测已安装的应用"),
		widget.NewLabel("2. 从列表中选择要重置的应用"),
		widget.NewLabel("3. 确保应用已关闭（运行中的应用不能重置）"),
		widget.NewLabel("4. 点击【重置选中】按钮开始重置流程"),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("重置内容", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("• 设备ID和唯一标识"),
		widget.NewLabel("• 账户登录记录和凭据"),
		widget.NewLabel("• 缓存数据和历史记录"),
		widget.NewLabel("注意：操作前会自动创建备份"),
	)

	dialog.ShowCustom("帮助信息", "关闭", helpContent, app.mainWindow)
}

// onAbout handles the about button click
func (app *App) onAbout() {
	// Create project homepage hyperlink
	projectURL, _ := url.Parse("https://github.com/whispin/Cursor_Windsurf_Reset")
	projectLink := widget.NewHyperlink("项目主页", projectURL)
	projectLink.Alignment = fyne.TextAlignCenter

	aboutContent := container.NewVBox(
		widget.NewLabelWithStyle("Cursor & Windsurf 重置工具", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("版本: 1.0.0"),
		widget.NewLabel("基于Go语言和Fyne框架开发"),
		projectLink,
		widget.NewSeparator(),
		widget.NewLabel("此工具用于重置Cursor和Windsurf应用的数据"),
		widget.NewLabel("包括设备ID、账户记录和缓存数据"),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("注意事项", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("请在使用前备份重要数据"),
		widget.NewLabel("使用须知：本软件及其相关文档仅用于教育、学习与评估目的"),
		widget.NewSeparator(),
		widget.NewLabel("不可用于任何商业/非法用途，开发者不承担一切法律责任。"),
	)

	dialog.ShowCustom("关于", "关闭", aboutContent, app.mainWindow)
}

// 修改App结构体log方法，支持日志级别
func (app *App) log(level, message string) {
	level = strings.ToUpper(level)
	app.guiLogger.WithLevel(parseLevel(level)).Msg(message)
}

// parseLevel辅助函数
func parseLevel(level string) zerolog.Level {
	switch level {
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Run starts the GUI application
func (app *App) Run() {
	app.mainWindow.ShowAndRun()
}

// GetMainWindow returns the main window
func (app *App) GetMainWindow() fyne.Window {
	return app.mainWindow
}

func (app *App) createAppListArea() *fyne.Container {
	// 垂直布局容器，将包含所有应用卡片
	appsContainer := container.NewVBox()

	// 确保appData已经被初始化
	if len(app.appData) == 0 {
		app.log("INFO", "appData为空，无法创建应用列表")

		// 尝试从配置中手动创建应用列表
		if app.config != nil && len(app.config.Applications) > 0 {
			app.log("INFO", fmt.Sprintf("尝试从配置创建应用列表，应用数: %d", len(app.config.Applications)))

			// 使用排序的应用名称
			appNames := make([]string, 0, len(app.config.Applications))
			for appName := range app.config.Applications {
				appNames = append(appNames, appName)
			}
			// 排序应用名称
			sort.Strings(appNames)

			// 重置应用列表
			app.appData = make([]AppInfo, 0)

			// 按排序后的名称添加应用
			for _, appName := range appNames {
				appConfig := app.config.Applications[appName]
				app.log("INFO", fmt.Sprintf("从配置处理应用: %s, 显示名称: %s", appName, appConfig.DisplayName))

				// 创建应用信息对象
				appInfo := AppInfo{
					Name:        appName,
					DisplayName: appConfig.DisplayName,
					Path:        "未知",
					Size:        "未知",
					Found:       false,
					Selected:    false,
				}

				app.appData = append(app.appData, appInfo)
			}
		}
	}

	// 确保appData不为空
	if len(app.appData) == 0 {
		app.log("INFO", "无法创建应用列表，appData仍为空")
		return container.NewVBox(widget.NewLabel("找不到应用"))
	}

	// 调试日志
	app.log("INFO", fmt.Sprintf("创建应用列表区域, 应用数: %d", len(app.appData)))
	for i, appInfo := range app.appData {
		app.log("INFO", fmt.Sprintf("应用列表项: 索引: %d, 显示名称: %s", i, appInfo.DisplayName))
	}

	// 计算已找到和可重置的应用数量
	foundCount := 0
	cleanableCount := 0
	for _, appInfo := range app.appData {
		if appInfo.Found {
			foundCount++
			if !appInfo.Running {
				cleanableCount++
			}
		}
	}

	// 创建状态文字
	statusText := fmt.Sprintf("已发现: %d  可重置: %d", foundCount, cleanableCount)

	// 列表标题
	listHeader := container.NewHBox(
		widget.NewLabelWithStyle("应用程序列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewLabelWithStyle(statusText, fyne.TextAlignTrailing, fyne.TextStyle{Italic: true}),
		app.selectAllCheck,
	)

	// 创建边框来分隔应用列表区域
	listBorder := canvas.NewRectangle(color.NRGBA{R: 50, G: 55, B: 65, A: 100})
	listBorder.StrokeWidth = 1
	listBorder.StrokeColor = color.NRGBA{R: 60, G: 70, B: 80, A: 150}

	// 遍历应用程序并创建卡片
	for i, appInfo := range app.appData {
		// 索引拷贝，避免闭包问题
		index := i

		// 复选框
		checkBox := widget.NewCheck("", func(checked bool) {
			app.selectedApps[index] = checked
			app.updateCleanButton()
			app.log("INFO", fmt.Sprintf("复选框状态变更: 索引: %d, 显示名称: %s, 选中: %v", index, appInfo.DisplayName, checked))
		})

		// 设置复选框的选中状态
		checkBox.SetChecked(app.selectedApps[index])

		// 如果应用未找到或正在运行，则禁用复选框
		if !appInfo.Found || appInfo.Running {
			checkBox.Disable()
		}

		// 状态图标 - 使用更明显的图标和颜色
		var statusIcon *widget.Icon
		var statusText string
		var statusColor color.Color

		if appInfo.Found {
			if appInfo.Running {
				statusIcon = widget.NewIcon(theme.MediaPlayIcon())
				statusText = "运行中"
				statusColor = color.NRGBA{R: 255, G: 180, B: 0, A: 255} // 橙黄色
			} else {
				statusIcon = widget.NewIcon(theme.ConfirmIcon())
				statusText = "可重置"
				statusColor = color.NRGBA{R: 50, G: 205, B: 50, A: 255} // 绿色
			}
		} else {
			statusIcon = widget.NewIcon(theme.ErrorIcon())
			statusText = "未找到"
			statusColor = color.NRGBA{R: 255, G: 70, B: 70, A: 255} // 红色
			checkBox.Disable()
		}

		// 创建状态指示器 - 减小尺寸
		statusIndicator := canvas.NewRectangle(statusColor)
		statusIndicator.SetMinSize(fyne.NewSize(3, 18)) // 进一步减小高度

		// 路径显示
		var pathText string
		if appInfo.Found {
			pathText = appInfo.Path
		} else {
			pathText = "N/A"
		}

		// 创建路径标签，确保更加醒目和清晰可见
		pathLabel := widget.NewLabel(fmt.Sprintf("路径: %s", pathText))

		// 直接使用普通标签，确保路径始终显示，不使用TextTruncate
		pathLabel.Alignment = fyne.TextAlignLeading
		pathLabel.TextStyle = fyne.TextStyle{
			Bold:   false,
			Italic: true,
		}

		// 状态标签
		statusLabel := widget.NewLabelWithStyle(statusText, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

		// 尺寸显示
		sizeContainer := container.NewHBox(
			widget.NewIcon(theme.StorageIcon()),
			widget.NewLabelWithStyle(appInfo.Size, fyne.TextAlignTrailing, fyne.TextStyle{}),
		)

		// 创建标题行
		titleRow := container.NewHBox(
			statusIndicator,
			checkBox,
			statusIcon,
			widget.NewLabelWithStyle(appInfo.DisplayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
			statusLabel,
		)

		// 创建路径行 - 使用独立框架突出显示路径
		pathFrame := canvas.NewRectangle(color.NRGBA{R: 40, G: 45, B: 55, A: 120})
		pathFrame.StrokeWidth = 1
		pathFrame.StrokeColor = color.NRGBA{R: 70, G: 80, B: 90, A: 150}

		// 路径和大小信息行
		pathRow := container.NewHBox(
			widget.NewIcon(theme.FolderIcon()),
			pathLabel,
			layout.NewSpacer(),
			sizeContainer,
		)

		// 创建一条分隔线使路径与标题分开
		separator := widget.NewSeparator()

		// 创建更紧凑的卡片内容
		cardContent := container.NewVBox(
			titleRow,
			separator,
			container.NewPadded(pathRow), // 使用内边距包装路径行，增加可见性
		)

		// 背景带有更明显的边框，增强卡片效果
		bg := canvas.NewRectangle(color.NRGBA{R: 45, G: 50, B: 60, A: 60})
		bg.StrokeWidth = 1
		bg.StrokeColor = color.NRGBA{R: 70, G: 80, B: 90, A: 120}

		// 使用Container.New创建一个自定义容器
		card := container.New(
			layout.NewMaxLayout(),
			bg,
			container.NewPadded(cardContent),
		)

		// 将卡片添加到容器
		appsContainer.Add(card)
	}

	// 创建滚动容器
	scrollContainer := container.NewScroll(appsContainer)

	// 减小应用列表区域的高度
	scrollContainer.SetMinSize(fyne.NewSize(0, 90))

	// 返回完整的应用列表区域，包含边框
	return container.New(
		layout.NewMaxLayout(),
		listBorder,
		container.NewBorder(
			listHeader,
			nil, nil, nil,
			scrollContainer,
		),
	)
}

// 刷新应用列表
func (app *App) refreshAppList() {
	// 如果是初始化阶段，不执行操作
	if app.mainWindow == nil || app.mainWindow.Content() == nil {
		app.log("INFO", "无法刷新应用列表 - 窗口未初始化")
		return
	}

	// 记录当前时间，用于性能分析
	startTime := time.Now()
	app.log("INFO", "正在刷新应用列表")

	// 重新创建应用列表区域
	appListArea := app.createAppListArea()

	// 更新主区域容器的内容
	if app.mainAreaContainer != nil {
		// mainAreaContainer是一个Border布局，其对象顺序为 [center, top, bottom, left, right]
		// 我们需要替换中央内容（第一个对象），同时保留底部控件
		app.mainAreaContainer.Objects[0] = appListArea
		app.mainAreaContainer.Refresh()
		app.log("INFO", "应用列表UI已刷新")
	} else {
		app.log("INFO", "主区域容器为nil，无法刷新UI")
	}

	// 更新重置按钮状态
	app.updateCleanButton()

	// 记录完成时间
	elapsedTime := time.Since(startTime)
	app.log("INFO", fmt.Sprintf("应用列表刷新完成，用时: %v", elapsedTime))
}
