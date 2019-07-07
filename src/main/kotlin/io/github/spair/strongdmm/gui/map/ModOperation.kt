package io.github.spair.strongdmm.gui.map

import io.github.spair.strongdmm.gui.map.select.SelectOperation
import io.github.spair.strongdmm.logic.map.Dmm
import io.github.spair.strongdmm.logic.map.Tile
import io.github.spair.strongdmm.logic.map.TileOperation

object ModOperation {

    fun copy(map: Dmm, x: Int, y: Int) {
        val pickedTiles = getPickedTiles()

        if (pickedTiles != null) {
            TileOperation.copy(pickedTiles)
        } else {
            TileOperation.copy(map.tile(x, y))
        }
    }

    fun cut(map: Dmm, x: Int, y: Int) {
        val pickedTiles = getPickedTiles()

        if (pickedTiles != null) {
            TileOperation.cut(map, pickedTiles)
        } else {
            TileOperation.cut(map, map.tile(x, y))
        }

        SelectOperation.depickArea()
        Frame.update(true)
    }

    fun paste(map: Dmm, x: Int, y: Int) {
        TileOperation.paste(map, x, y) {
            SelectOperation.pickArea(it)
        }
        Frame.update(true)
    }

    fun delete(map: Dmm, x: Int, y: Int) {
        val pickedTiles = getPickedTiles()

        if (pickedTiles != null) {
            TileOperation.delete(map, pickedTiles)
        } else {
            TileOperation.delete(map, map.tile(x, y))
        }

        SelectOperation.depickArea()
        Frame.update(true)
    }

    private fun Dmm.tile(x: Int, y: Int): Tile = getTile(x, y)!!
    private fun getPickedTiles(): List<Tile>? = SelectOperation.getPickedTiles()?.takeIf { it.isNotEmpty() }
}
