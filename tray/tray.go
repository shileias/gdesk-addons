package tray

/*
#include <windows.h>
*/
import "C"
import (
	"log"
	"os"
	"sync"

	"github.com/getlantern/systray"
	windows "github.com/shileias/gdesk/command/windows"
)

var (
	instance *Tray
	once     sync.Once
)

type TrayMenuItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Checked     bool   `json:"checked"`
	Enabled     bool   `json:"enabled"`
	IsSeparator bool   `json:"is_separator"`
	item        *systray.MenuItem
}

type Tray struct {
	mainWindow *windows.WindowImpl
	title      string
	icon       []byte
	menuItems  map[string]*TrayMenuItem
	itemsOrder []*TrayMenuItem
	mu         sync.Mutex
	onClick    func(string)
	enabled    bool
}

func NewTray() *Tray {
	once.Do(func() {
		instance = &Tray{
			mainWindow: nil,
			menuItems:  make(map[string]*TrayMenuItem),
			enabled:    true,
		}
	})
	return instance
}

func (t *Tray) SetTitle(title string) {
	t.title = title
}

func (t *Tray) SetIcon(iconBytes []byte) {
	t.icon = iconBytes
}

func (t *Tray) AddItem(id string, title string, checked bool) *TrayMenuItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &TrayMenuItem{
		ID:          id,
		Title:       title,
		Checked:     checked,
		Enabled:     true,
		IsSeparator: false,
	}

	t.menuItems[id] = item
	t.itemsOrder = append(t.itemsOrder, item)

	return item
}

func (t *Tray) AddSeparator() {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &TrayMenuItem{
		IsSeparator: true,
	}
	t.itemsOrder = append(t.itemsOrder, item)
}

func (t *Tray) SetChecked(id string, checked bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if item, ok := t.menuItems[id]; ok {
		item.Checked = checked
		if checked {
			item.item.Check()
		} else {
			item.item.Uncheck()
		}
	}
}

func (t *Tray) SetEnabled(id string, enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if item, ok := t.menuItems[id]; ok {
		item.Enabled = enabled
		if enabled {
			item.item.Enable()
		} else {
			item.item.Disable()
		}
	}
}

func (t *Tray) RemoveItem(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if item, ok := t.menuItems[id]; ok {
		item.item.Hide()
		delete(t.menuItems, id)
	}
}

func (t *Tray) Start() {
	go systray.Run(func() {
		if t.icon != nil {
			systray.SetIcon(t.icon)
		}
		if t.title != "" {
			systray.SetTitle(t.title)
			systray.SetTooltip(t.title)
		}

		t.mu.Lock()
		for _, item := range t.itemsOrder {
			if item.IsSeparator {
				systray.AddSeparator()
				continue
			}
			item.item = systray.AddMenuItem(item.Title, "")
			if item.Checked {
				item.item.Check()
			} else {
				item.item.Uncheck()
			}
			go func(it *TrayMenuItem) {
				for range it.item.ClickedCh {
					if t.onClick != nil {
						t.onClick(it.ID)
					}
				}
			}(item)
		}
		t.mu.Unlock()
	}, func() {
	})
}

func (t *Tray) Stop() {
	systray.Quit()
}

func (t *Tray) SetOnClick(handler func(string)) {
	t.onClick = handler
}

func (t *Tray) GetMenuItems() []TrayMenuItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	var items []TrayMenuItem
	for _, item := range t.menuItems {
		items = append(items, TrayMenuItem{
			ID:      item.ID,
			Title:   item.Title,
			Checked: item.Checked,
			Enabled: item.Enabled,
		})
	}
	return items
}

func (t *Tray) ToggleMainWindow() {
	app := windows.GetApp()
	if app == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mainWindow == nil {
		t.mainWindow = app.MainWindow()
		if t.mainWindow == nil {
			return
		}
	}
	visible := t.mainWindow.IsVisible()
	log.Printf("[tray] ToggleMainWindow: IsVisible = %v\n", visible)
	if visible {
		t.mainWindow.Hide()
	} else {
		ShowWindowForce(t.mainWindow)
	}
	log.Printf("[tray] After toggle: IsVisible = %v\n", t.mainWindow.IsVisible())
}

func (t *Tray) ShowMainWindow() {
	app := windows.GetApp()
	if app == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mainWindow == nil {
		t.mainWindow = app.MainWindow()
		if t.mainWindow == nil {
			return
		}
	}
	log.Printf("[tray] ShowMainWindow: before IsVisible = %v\n", t.mainWindow.IsVisible())
	ShowWindowForce(t.mainWindow)
	log.Printf("[tray] ShowMainWindow: after IsVisible = %v\n", t.mainWindow.IsVisible())
}

func (t *Tray) HideMainWindow() {
	app := windows.GetApp()
	if app == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.mainWindow == nil {
		t.mainWindow = app.MainWindow()
		if t.mainWindow == nil {
			return
		}
	}
	log.Printf("[tray] HideMainWindow: before IsVisible = %v\n", t.mainWindow.IsVisible())
	t.mainWindow.Hide()
	log.Printf("[tray] HideMainWindow: after IsVisible = %v\n", t.mainWindow.IsVisible())
}

func ShowWindowForce(window *windows.WindowImpl) {
	hwnd := C.HWND(window.NativeHandle())
	if hwnd == nil {
		return
	}
	C.ShowWindow(hwnd, C.SW_SHOW)
	C.SetForegroundWindow(hwnd)
	C.UpdateWindow(hwnd)
}

func (t *Tray) Exit() {
	app := windows.GetApp()
	if app != nil {
		app.Destroy()
	}
	t.Stop()
	os.Exit(0)
}

type WindowState struct {
	IsVisible   bool `json:"is_visible"`
	IsMinimized bool `json:"is_minimized"`
	IsMaximized bool `json:"is_maximized"`
}

func GetWindowState(window *windows.WindowImpl) *WindowState {
	if instance != nil && window != nil {
		return &WindowState{
			IsVisible:   window.IsVisible(),
			IsMinimized: window.IsMinimised(),
			IsMaximized: window.IsMaximised(),
		}
	}
	return &WindowState{
		IsVisible:   true,
		IsMinimized: false,
		IsMaximized: false,
	}
}

func ToggleMainWindow() {
	if instance == nil {
		return
	}
	instance.ToggleMainWindow()
}

func ShowMainWindow() {
	if instance == nil {
		return
	}
	instance.ShowMainWindow()
}

func HideMainWindow() {
	if instance == nil {
		return
	}
	instance.HideMainWindow()
}

func autoInit() {
	tray := NewTray()
	tray.SetTitle("GDesk")

	iconPath := "./resources/app.ico"
	iconBytes, err := os.ReadFile(iconPath)
	if err == nil {
		tray.SetIcon(iconBytes)
	}

	tray.AddItem("toggle", "显示/隐藏主窗口", false)
	tray.AddSeparator()
	tray.AddItem("quit", "退出应用", false)

	tray.SetOnClick(func(id string) {
		switch id {
		case "toggle":
			tray.ToggleMainWindow()
		case "quit":
			tray.Exit()
		}
	})

	tray.Start()
}

func init() {
	windows.RegisterPluginFunctions("plugins/tray",
		"ToggleMainWindow", ToggleMainWindow,
		"ShowMainWindow", ShowMainWindow,
		"HideMainWindow", HideMainWindow,
		"GetWindowState", GetWindowState,
	)

	go func() {
		autoInit()
	}()
}
