package com.brackistar.gamemasternotes.core.design

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.Immutable
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

private val LightColors = lightColorScheme(
    primary = Color(0xFF087EA4),
    onPrimary = Color.White,
    primaryContainer = Color(0xFFD7F5FF),
    onPrimaryContainer = Color(0xFF073147),
    secondary = Color(0xFF304A59),
    background = Color(0xFFF3F6F8),
    surface = Color(0xFFFFFFFF),
    surfaceVariant = Color(0xFFEEF3F6),
    outline = Color(0xFFC9D5DC),
    onBackground = Color(0xFF142C3A),
    onSurface = Color(0xFF142C3A),
    onSurfaceVariant = Color(0xFF304A59),
    error = Color(0xFFB42318),
    errorContainer = Color(0xFFFDE8E7),
)

private val DarkColors = darkColorScheme(
    primary = Color(0xFF5CC9ED),
    onPrimary = Color(0xFF002A38),
    primaryContainer = Color(0xFF123F50),
    onPrimaryContainer = Color(0xFFD7F5FF),
    secondary = Color(0xFFC6D6DE),
    background = Color(0xFF0D171D),
    surface = Color(0xFF142229),
    surfaceVariant = Color(0xFF1C2D36),
    outline = Color(0xFF40545F),
    onBackground = Color(0xFFE8F2F6),
    onSurface = Color(0xFFE8F2F6),
    onSurfaceVariant = Color(0xFFC6D6DE),
    error = Color(0xFFFFB4AB),
    errorContainer = Color(0xFF5A1A18),
)

@Immutable
data class WayfinderColors(
    val navigationContainer: Color,
    val onNavigationContainer: Color,
    val navigationSelected: Color,
    val onNavigationSelected: Color,
)

private val LightWayfinderColors = WayfinderColors(
    navigationContainer = Color(0xFF102333),
    onNavigationContainer = Color(0xFFE2ECF2),
    navigationSelected = Color(0xFFD7F5FF),
    onNavigationSelected = Color(0xFF073147),
)

private val DarkWayfinderColors = WayfinderColors(
    navigationContainer = Color(0xFF08141D),
    onNavigationContainer = Color(0xFFE6F1F7),
    navigationSelected = Color(0xFF123F50),
    onNavigationSelected = Color(0xFFD7F5FF),
)

private val LocalWayfinderColors = staticCompositionLocalOf { LightWayfinderColors }

object WayfinderTheme {
    val colors: WayfinderColors
        @Composable
        @ReadOnlyComposable
        get() = LocalWayfinderColors.current
}

private val WayfinderTypography = Typography(
    headlineLarge = TextStyle(fontSize = 28.sp, lineHeight = 36.sp, fontWeight = FontWeight.Bold),
    headlineMedium = TextStyle(fontSize = 24.sp, lineHeight = 32.sp, fontWeight = FontWeight.Bold),
    titleLarge = TextStyle(fontSize = 20.sp, lineHeight = 28.sp, fontWeight = FontWeight.Bold),
    titleMedium = TextStyle(fontSize = 16.sp, lineHeight = 24.sp, fontWeight = FontWeight.SemiBold),
    bodyLarge = TextStyle(fontSize = 16.sp, lineHeight = 26.sp),
    bodyMedium = TextStyle(fontSize = 14.sp, lineHeight = 22.sp),
    labelLarge = TextStyle(fontSize = 14.sp, lineHeight = 20.sp, fontWeight = FontWeight.SemiBold),
    labelMedium = TextStyle(fontSize = 12.sp, lineHeight = 16.sp, fontWeight = FontWeight.SemiBold),
)

@Composable
fun GameMasterNotesTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    CompositionLocalProvider(
        LocalWayfinderColors provides if (darkTheme) DarkWayfinderColors else LightWayfinderColors,
    ) {
        MaterialTheme(
            colorScheme = if (darkTheme) DarkColors else LightColors,
            typography = WayfinderTypography,
        ) {
            Surface(
                modifier = Modifier.fillMaxSize(),
                color = MaterialTheme.colorScheme.background,
                contentColor = MaterialTheme.colorScheme.onBackground,
                content = content,
            )
        }
    }
}
