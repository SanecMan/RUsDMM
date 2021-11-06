package pmap

import (
	"fmt"

	"github.com/SpaiR/imgui-go"
)

func (p *PaneMap) showStatusPanel() {
	if p.canvasState.HoverOutOfBounds() {
		imgui.Text("[out of bounds]")
	} else {
		hoveredTiles := p.canvasState.HoveredTile()
		imgui.Text(fmt.Sprintf("[X:%03d Y:%03d]", hoveredTiles.X, hoveredTiles.Y))
	}

	if hoveredInstance := p.canvasState.HoveredInstance(); hoveredInstance != nil {
		imgui.SameLine()
		imgui.Text(hoveredInstance.Prefab().Path())
	}
}
