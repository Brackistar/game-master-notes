package com.brackistar.gamemasternotes

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import com.brackistar.gamemasternotes.core.design.GameMasterNotesTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            GameMasterNotesTheme {
                GameMasterNotesApp()
            }
        }
    }
}
