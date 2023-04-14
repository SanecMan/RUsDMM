package psettings

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"time"

	"sdmm/internal/app/render/bucket/level/chunk/unit"
	"sdmm/internal/app/ui/cpwsarea/wsmap/pmap/canvas"
	appdialog "sdmm/internal/app/ui/dialog"
	"sdmm/internal/dmapi/dmmap"
	"sdmm/internal/imguiext"
	"sdmm/internal/imguiext/icon"
	"sdmm/internal/imguiext/style"
	w "sdmm/internal/imguiext/widget"
	"sdmm/internal/util"

	"github.com/SpaiR/imgui-go"
	"github.com/sqweek/dialog"
)

type sessionScreenshot struct {
	saving bool
}

func (p *Panel) showScreenshot() {
	if imgui.CollapsingHeader("Screenshot") {
		if imgui.Button(icon.FolderOpen) {
			p.selectScreenshotDir()
		}

		imguiext.SetItemHoveredTooltip("Screenshot Folder")

		imgui.SameLine()

		imgui.SetNextItemWidth(-1)
		imgui.InputText("##screenshot_dir", &cfg.ScreenshotDir)

		var createBtnLabel string
		if p.sessionScreenshot.saving {
			createBtnLabel = "Creating" + []string{".", "..", "...", "...."}[int(imgui.Time()/.25)&3] + "###create"
		} else {
			createBtnLabel = icon.Save + "Create###create"
		}

		w.Layout{
			w.Disabled(p.sessionScreenshot.saving,
				w.Button(createBtnLabel, p.createScreenshot).
					Size(imgui.Vec2{X: -1}).
					Style(style.ButtonGreen{}),
			),
		}.Build()
	}
}

func (p *Panel) createScreenshot() {
	p.sessionScreenshot.saving = true

	width, height := p.editor.Dmm().MaxX*dmmap.WorldIconSize, p.editor.Dmm().MaxY*dmmap.WorldIconSize

	c := canvas.New()
	c.ClearColor = canvas.Color{} // Empty clear color with no alpha
	c.Render().Camera.Level = p.editor.ActiveLevel()
	c.Render().SetUnitProcessor(p)
	for level := 1; level <= p.editor.ActiveLevel(); level++ {
		c.Render().UpdateBucket(p.editor.Dmm(), level) // Prepare for render all available levels
	}
	c.Process(imgui.Vec2{X: float32(width), Y: float32(height)})
	c.Dispose()

	var pixels = c.ReadPixels()

	go func() {
		if err := saveScreenshot(pixels, width, height); err != nil {
			appdialog.Open(appdialog.TypeInformation{
				Title:       "Error: Screenshot Creation",
				Information: fmt.Sprint("Unable to create screenshot:", err),
			})
		}
		p.sessionScreenshot.saving = false
	}()
}

func (p *Panel) selectScreenshotDir() {
	if dir, err := dialog.
		Directory().
		Title("Screenshot Directory").
		SetStartDir(cfg.ScreenshotDir).
		Browse(); err == nil {
		log.Println("[psettings] screenshot directory selected:", dir)
		cfg.ScreenshotDir = dir
	}
}

func (p *Panel) ProcessUnit(u unit.Unit) bool {
	return p.app.PathsFilter().IsVisiblePath(u.Instance().Prefab().Path())
}

func saveScreenshot(pixels []byte, w, h int) error {
	if err := os.MkdirAll(cfg.ScreenshotDir, os.ModeDir); err != nil {
		log.Println("[psettings] unable to create screenshot directory:", err)
		return err
	}

	out, _ := os.Create(cfg.ScreenshotDir + "/" + time.Now().Format(util.TimeFormat) + ".png")
	defer out.Close()

	return png.Encode(out, util.PixelsToRGBA(pixels, w, h))
}