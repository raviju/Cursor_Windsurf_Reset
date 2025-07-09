# Cursor & Windsurf 重置工具

用于重置 Cursor 和 Windsurf 应用数据的高级工具，支持最新版Cursor 和 Windsurf，

## ✨ 界面截图

![界面截图](https://github.com/whispin/Cursor_Windsurf_Reset/blob/main/screenshot/homepage_zh.jpg?raw=true)
## ✨ 功能特性

### 🎯 核心功能
- **支持最新版Cursor 和 Windsurf**：Cursor 1.2.1,Windsurf 1.10.7
- **智能重置**：自动检测并重置 Cursor 和 Windsurf 的设备ID、会话数据和缓存
- **双界面支持**：提供现代化图形界面和功能完整的命令行界面
- **跨平台兼容**：支持 Windows、macOS 和 Linux 系统
- **安全备份**：重置前自动创建数据备份，支持一键恢复

## 📦 安装说明
### 方式一：下载预编译版本
1. 访问 [Releases 页面](https://github.com/whispin/Cursor_Windsurf_Reset/releases)
2. 下载适合您系统的版本：
   - Windows: `Cursor_Windsurf_Reset-windows.exe`
   - macOS: `Cursor_Windsurf_Reset-macos`
   - Linux: `Cursor_Windsurf_Reset-linux`
3. 双击运行（Windows）或在终端中执行
#### 操作步骤
1. 启动应用后，工具会自动检测已安装的应用
2. 选择要重置的应用（Cursor、Windsurf 或全部）
3. 点击"开始重置"按钮
4. 确认操作并等待完成
5. 查看操作结果和备份位置

## 🛠️ 开发说明

### 技术栈
- **语言**：Go 1.21+
- **GUI框架**：Fyne v2
### 项目结构
```
Cursor_Windsurf_Reset-go/
├── main.go                 # 主程序入口
├── cleaner/
│   └── engine.go          # 清理引擎核心逻辑
├── config/
│   └── config.go          # 配置管理
├── gui/
│   ├── app.go             # GUI应用主逻辑
│   ├── theme.go           # 主题定义
│   └── resources.go       # 资源文件
├── reset_config.json    # 默认配置文件
├── go.mod                 # Go模块定义
└── README.md              # 项目说明
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 贡献指南
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📞 支持

- **GitHub Issues**：[提交问题](https://github.com/whispin/Cursor_Windsurf_Reset/issues)
- **项目主页**：[https://github.com/whispin/Cursor_Windsurf_Reset](https://github.com/whispin/Cursor_Windsurf_Reset)

## ⚠️ 免责声明

本工具仅供学习和研究使用。使用本工具时请：

1. **备份数据**：使用前请备份重要数据
2. **遵守条款**：遵守相关应用程序的服务条款
3. **自担风险**：使用本工具的风险由用户自行承担
4. **合法使用**：确保使用方式符合当地法律法规

---
