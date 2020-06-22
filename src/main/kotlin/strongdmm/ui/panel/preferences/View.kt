package strongdmm.ui.panel.preferences

import imgui.ImGui.*
import imgui.ImInt
import imgui.enums.ImGuiCond
import imgui.enums.ImGuiMouseButton
import imgui.enums.ImGuiMouseCursor
import imgui.enums.ImGuiStyleVar
import strongdmm.service.preferences.prefs.PreferenceBoolean
import strongdmm.service.preferences.prefs.PreferenceEnum
import strongdmm.service.preferences.prefs.PreferenceInteger
import strongdmm.util.imgui.*
import strongdmm.util.imgui.inputIntClamp
import strongdmm.window.Window

class View(
    private val state: State
) {
    companion object {
        private val width: Float
            get() = 450f * Window.pointSize
        private val height: Float
            get() = 500f * Window.pointSize

        private const val TITLE: String = "Preferences"

        private val toggleButtonPadding: Float
            get() = Window.pointSize

        private val optionsIndent: Float
            get() = 10f * Window.pointSize
    }

    lateinit var viewController: ViewController

    private val imInt: ImInt = ImInt()

    fun process() {
        if (state.isDoOpen) {
            state.checkOpenStatus = true
            openPopup(TITLE)
            viewController.blockApplication()
            state.isDoOpen = false
        }

        ImGuiUtil.setNextWindowCentered(width, height, ImGuiCond.Appearing)

        popupModal(TITLE, state.isOpened) {
            state.preferencesByGroups.forEach { (groupName, preferences) ->
                textColored(1f, .84f, 0f, .75f, groupName)
                separator()

                withIndent(optionsIndent) {
                    preferences.forEach { pref ->
                        when (pref) {
                            is PreferenceInteger -> showInputIntOption(pref)
                            is PreferenceEnum -> showSelectOption(pref)
                            is PreferenceBoolean -> showToggleOption(pref)
                        }

                        newLine()
                    }
                }
            }
        }

        viewController.checkOpenStatus()
    }

    private fun showInputIntOption(pref: PreferenceInteger) {
        textWrapped(pref.getHeader())

        pushTextWrapPos()
        textDisabled(pref.getDesc())
        popTextWrapPos()

        imInt.set(pref.getValue().data)
        if (inputIntClamp(pref.getLabel(), imInt, pref.getMin(), pref.getMax(), pref.getStep(), pref.getStepFast())) {
            pref.getValue().data = imInt.get()
            viewController.savePreferences()
        }
    }

    private fun showSelectOption(pref: PreferenceEnum) {
        textWrapped(pref.getHeader())

        pushTextWrapPos()
        textDisabled(pref.getDesc())
        popTextWrapPos()

        combo(pref.getLabel(), pref.getReadableName()) {
            pref.getEnums().forEach { mode ->
                selectable(mode.getReadableName(), pref == mode) {
                    viewController.doSelectOption(mode, pref)
                }
            }
        }
    }

    private fun showToggleOption(pref: PreferenceBoolean) {
        textWrapped(pref.getHeader())

        withStyleVar(ImGuiStyleVar.FramePadding, toggleButtonPadding, toggleButtonPadding) {
            if (checkbox(pref.getLabel(), pref.getValue().data)) {
                viewController.doToggleOption(pref)
            }
        }

        sameLine()

        pushTextWrapPos()
        textDisabled(pref.getDesc())
        popTextWrapPos()

        if (isItemHovered()) {
            setMouseCursor(ImGuiMouseCursor.Hand)
        }

        if (isItemClicked(ImGuiMouseButton.Left)) {
            viewController.doToggleOption(pref)
        }
    }
}
