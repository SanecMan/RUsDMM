package strongdmm.service.tool.tile

import strongdmm.byond.dmm.MapPos
import strongdmm.byond.dmm.TileItem
import strongdmm.event.EventBus
import strongdmm.event.service.TriggerActionService
import strongdmm.event.service.TriggerCanvasService
import strongdmm.event.service.TriggerFrameService
import strongdmm.event.service.TriggerMapHolderService
import strongdmm.service.action.undoable.MultiAction
import strongdmm.service.action.undoable.ReplaceTileAction
import strongdmm.service.action.undoable.Undoable
import strongdmm.service.tool.Tool

class TileAddTool : Tool() {
    private val dirtyTiles: MutableSet<MapPos> = mutableSetOf()
    private val reverseActions: MutableList<Undoable> = mutableListOf()

    private var selectedTileItem: TileItem? = null

    override fun onStart(mapPos: MapPos) {
        isActive = selectedTileItem != null

        if (isActive && dirtyTiles.add(mapPos)) {
            addTileItem(mapPos)
        }
    }

    override fun onStop() {
        flushReverseActions()
        reset()
    }

    override fun onMapPosChanged(mapPos: MapPos) {
        if (dirtyTiles.add(mapPos)) {
            addTileItem(mapPos)
        }
    }

    override fun onTileItemSwitch(tileItem: TileItem?) {
        selectedTileItem = tileItem
    }

    override fun reset() {
        isActive = false
        dirtyTiles.clear()
        reverseActions.clear()
        EventBus.post(TriggerCanvasService.ResetSelectedTiles())
    }

    override fun destroy() {
        reset()
        selectedTileItem = null
    }

    private fun addTileItem(pos: MapPos) {
        EventBus.post(TriggerMapHolderService.FetchSelectedMap { selectedMap ->
            val tile = selectedMap.getTile(pos.x, pos.y, selectedMap.zSelected)

            reverseActions.add(ReplaceTileAction(tile) {
                tile.addTileItem(selectedTileItem!!)
            })

            EventBus.post(TriggerCanvasService.SelectTiles(dirtyTiles))
            EventBus.post(TriggerFrameService.RefreshFrame())
        })
    }

    private fun flushReverseActions() {
        if (reverseActions.isEmpty()) {
            return
        }

        EventBus.post(TriggerActionService.QueueUndoable(MultiAction(reverseActions.toList())))
    }
}
